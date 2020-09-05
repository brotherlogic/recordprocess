package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	getter getter
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

func (p prodGetter) update(ctx context.Context, instanceID int32, cat pbrc.ReleaseMetadata_Category, reason string, ncount int32) error {
	conn, err := p.dial(ctx, "recordcollection")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	_, err = client.UpdateRecord(ctx, &pbrc.UpdateRecordRequest{Reason: reason, Requestor: "recordprocess", Update: &pbrc.Record{Release: &gdpb.Release{InstanceId: instanceID}, Metadata: &pbrc.ReleaseMetadata{Category: cat, SaleAttempts: ncount}}})
	if err != nil {
		return err
	}
	return nil
}

// Init builds the server
func Init() *Server {
	s := &Server{
		GoServer: &goserver.GoServer{},
	}
	s.getter = &prodGetter{s.FDialServer}
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

//func (s *Server) saveScores(ctx context.Context) {
//	s.KSclient.Save(ctx, KEY, s.scores)
//}

func (s *Server) saveConfig(ctx context.Context, config *pb.Config) error {
	s.setVarz(config)
	return s.KSclient.Save(ctx, CONFIG, config)
}

func (s *Server) readScores(ctx context.Context) (*pb.Scores, error) {
	scores := &pb.Scores{}
	data, _, err := s.KSclient.Read(ctx, KEY, scores)

	if err != nil {
		return nil, err
	}

	return data.(*pb.Scores), nil
}

func (s *Server) readConfig(ctx context.Context) (*pb.Config, error) {
	config := &pb.Config{}
	data, _, err := s.KSclient.Read(ctx, CONFIG, config)

	if err != nil {
		return nil, err
	}

	config = data.(*pb.Config)

	// Ensure that we have recent updates on everything
	ids := []int32{}
	for id, next := range config.GetNextUpdateTime() {
		if time.Unix(next, 0).Sub(time.Now()) > time.Hour*24*3 {
			ids = append(ids, id)
		}
	}

	for _, id := range ids {
		config.NextUpdateTime[id] = time.Now().Unix()
	}

	s.Log(fmt.Sprintf("Updated %v, -> %v", len(ids), s.saveConfig(ctx, config)))

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
	return []*pbg.State{}
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

func (s *Server) procLoop() {
	for true {
		s.runElectLoop()

		//Wait 1 minute between runs
		time.Sleep(time.Minute)
	}
}

func (s *Server) runElectLoop() {
	cf, err := s.Elect()
	defer cf()

	if err == nil {
		s.runLoop()
	} else {
		s.Log(fmt.Sprintf("Unable to elect: %v", err))
	}
}

func (s *Server) setVarz(config *pb.Config) {
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

}

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
	code := status.Convert(err).Code()
	if code != codes.OK && code != codes.DeadlineExceeded && code != codes.NotFound {
		log.Fatalf("Unable to read config: %v", err)
	}
	server.setVarz(config)

	go server.procLoop()

	server.Serve()
}
