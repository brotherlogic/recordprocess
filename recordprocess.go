package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	gdpb "github.com/brotherlogic/godiscogs"
	pbg "github.com/brotherlogic/goserver/proto"
	"github.com/brotherlogic/goserver/utils"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	rcpb "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordprocess/proto"
)

const (
	// KEY used to save scores
	KEY = "github.com/brotherlogic/recordprocess/scores"

	// CONFIG where we store the config
	CONFIG = "github.com/brotherlogic/recordprocess/config"
)

//Server main server type
type Server struct {
	*goserver.GoServer
	config           *pb.Config
	getter           getter
	lastProc         time.Time
	lastCount        int64
	lastProcDuration time.Duration
	scores           *pb.Scores
	updates          int64
	recordsInUpdate  int64
	lastUpdate       int64
	updateCount      int
	configMutex      *sync.Mutex
}

type prodGetter struct {
	dial func(ctx context.Context, server string) (*grpc.ClientConn, error)
}

func (p prodGetter) getRecords(ctx context.Context, t int64, count int) ([]int32, error) {
	conn, err := p.dial(ctx, "recordcollection")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	t2 := t
	if count == 0 {
		t2 = 0
	}
	req := &pbrc.QueryRecordsRequest{Query: &pbrc.QueryRecordsRequest_UpdateTime{t2}}
	client := pbrc.NewRecordCollectionServiceClient(conn)
	resp, err := client.QueryRecords(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetInstanceIds(), nil
}

func (p prodGetter) getRecord(ctx context.Context, instanceID int32) (*pbrc.Record, error) {
	conn, err := p.dial(ctx, "recordcollection")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	resp, err := client.GetRecord(ctx, &pbrc.GetRecordRequest{InstanceId: instanceID})
	if err != nil {
		return nil, err
	}

	return resp.GetRecord(), nil
}

func (p prodGetter) update(ctx context.Context, instanceID int32, cat pbrc.ReleaseMetadata_Category, reason string) error {
	conn, err := p.dial(ctx, "recordcollection")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	_, err = client.UpdateRecord(ctx, &pbrc.UpdateRecordRequest{Reason: reason, Requestor: "recordprocess", Update: &pbrc.Record{Release: &gdpb.Release{InstanceId: instanceID}, Metadata: &pbrc.ReleaseMetadata{Category: cat}}})
	if err != nil {
		return err
	}
	return nil
}

// Init builds the server
func Init() *Server {
	s := &Server{
		GoServer:    &goserver.GoServer{},
		configMutex: &sync.Mutex{},
	}
	s.getter = &prodGetter{s.FDialServer}
	s.config = &pb.Config{}
	return s
}

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterScoreServiceServer(server, s)
	rcpb.RegisterClientUpdateServiceServer(server, s)
}

// ReportHealth alerts if we're not healthy
func (s *Server) ReportHealth() bool {
	return true
}

func (s *Server) saveScores(ctx context.Context) {
	s.configMutex.Lock()
	defer s.configMutex.Unlock()
	s.KSclient.Save(ctx, KEY, s.scores)
}

func (s *Server) saveConfig(ctx context.Context, config *pb.Config) error {
	return s.KSclient.Save(ctx, CONFIG, config)
}

func (s *Server) readScores(ctx context.Context) error {
	scores := &pb.Scores{}
	data, _, err := s.KSclient.Read(ctx, KEY, scores)

	if err != nil {
		return err
	}

	s.scores = data.(*pb.Scores)
	return nil
}

func (s *Server) readConfig(ctx context.Context) (*pb.Config, error) {
	config := &pb.Config{}
	data, _, err := s.KSclient.Read(ctx, CONFIG, config)

	if err != nil {
		return nil, err
	}

	config = data.(*pb.Config)

	return config, nil
}

func (s *Server) updateTime(ctx context.Context, iid int32, ti int64) error {
	config, err := s.readConfig(ctx)
	if err != nil {
		return err
	}

	config.NextUpdateTime[iid] = ti
	return s.saveConfig(ctx, config)
}

// Shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

// Mote promotes/demotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {

	numScores := int64(0)
	if s.scores != nil {
		numScores = int64(len(s.scores.Scores))
	}
	s.configMutex.Lock()
	defer s.configMutex.Unlock()
	return []*pbg.State{
		&pbg.State{Key: "queue_size", Value: int64(len(s.config.GetNextUpdateTime()))},
		&pbg.State{Key: "last_run_time", TimeValue: s.config.GetLastRunTime()},
		&pbg.State{Key: "size_scores", Value: int64(proto.Size(s.scores))},
		&pbg.State{Key: "num_scores", Value: numScores},
		&pbg.State{Key: "update_count", Value: int64(s.updateCount)},
		&pbg.State{Key: "last_proc", TimeValue: s.lastProc.Unix(), Value: s.lastCount},
		&pbg.State{Key: "last_proc_time", Text: fmt.Sprintf("%v", s.lastProcDuration)},
		&pbg.State{Key: "updates", Value: s.updates},
		&pbg.State{Key: "records_in_update", Value: s.recordsInUpdate},
		&pbg.State{Key: "last_record", Value: s.lastUpdate},
	}
}

var (
	size = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "recordprocess_queue",
		Help: "The amount of potential salve value",
	})

	nextUpdateTime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "recordprocess_nextupdatetime",
		Help: "The time of the next recordprocess update",
	})

	lastUpdateTime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "recordprocess_lastupdatetime",
		Help: "The time of the last recordprocess update",
	})
)

func main() {
	var quiet = flag.Bool("quiet", false, "Show all output")
	flag.Parse()

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	server := Init()
	server.PrepServer()
	server.Register = server

	err := server.RegisterServerV2("recordprocess", false, true)
	if err != nil {
		return
	}

	ctx, cancel := utils.ManualContext("recordproc", "recordproc", time.Minute, false)
	config, err := server.readConfig(ctx)
	cancel()
	if err != nil {
		log.Fatalf("Unable to read config")
	}

	size.Set(float64(len(config.GetNextUpdateTime())))
	min := time.Now().Unix()
	max := int64(0)
	for _, val := range config.GetNextUpdateTime() {
		if val < min {
			min = val
		}

		if val > max {
			max = val
		}
	}
	nextUpdateTime.Set(float64(min))
	lastUpdateTime.Set(float64(max))

	server.Serve()
}
