package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/goserver/utils"
	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbg "github.com/brotherlogic/goserver/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordprocess/proto"
)

const (
	// KEY used to save scores
	KEY = "github.com/brotherlogic/recordprocess/scores"
)

//Server main server type
type Server struct {
	*goserver.GoServer
	getter           getter
	lastProc         time.Time
	lastCount        int64
	lastProcDuration time.Duration
	scores           *pb.Scores
	updates          int64
	recordsInUpdate  int64
}

type prodGetter struct {
	getIP func(string) (string, int32, error)
}

func (p prodGetter) getRecords(ctx context.Context) ([]*pbrc.Record, error) {
	ip, port, err := p.getIP("recordcollection")
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	req := &pbrc.GetRecordsRequest{Filter: &pbrc.Record{}}
	resp, err := client.GetRecords(ctx, req, grpc.MaxCallRecvMsgSize(1024*1024*1024))
	if err != nil {
		return nil, err
	}

	return resp.GetRecords(), nil
}

func (p prodGetter) update(ctx context.Context, r *pbrc.Record) error {
	ip, port, err := p.getIP("recordcollection")
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	_, err = client.UpdateRecord(ctx, &pbrc.UpdateRecordRequest{Requestor: "recordprocess", Update: r})
	if err != nil {
		return err
	}
	return nil
}

func (p prodGetter) moveToSold(ctx context.Context, r *pbrc.Record) {
	ip, port, err := p.getIP("recordcollection")
	if err != nil {
		return
	}

	conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		return
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	r.GetMetadata().Category = pbrc.ReleaseMetadata_SOLD
	client.UpdateRecord(ctx, &pbrc.UpdateRecordRequest{Requestor: "recordprocess", Update: r})
}

// Init builds the server
func Init() *Server {
	s := &Server{GoServer: &goserver.GoServer{}}
	s.getter = &prodGetter{getIP: utils.Resolve}
	s.GoServer.KSclient = *keystoreclient.GetClient(s.GetIP)
	return s
}

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterScoreServiceServer(server, s)
}

// ReportHealth alerts if we're not healthy
func (s *Server) ReportHealth() bool {
	return true
}

func (s *Server) saveScores(ctx context.Context) {
	s.KSclient.Save(ctx, KEY, s.scores)
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

// Shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.saveScores(ctx)
	return nil
}

// Mote promotes/demotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	if master {
		err := s.readScores(ctx)
		return err
	}

	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	return []*pbg.State{
		&pbg.State{Key: "last_proc", TimeValue: s.lastProc.Unix(), Value: s.lastCount},
		&pbg.State{Key: "last_proc_time", Text: fmt.Sprintf("%v", s.lastProcDuration)},
		&pbg.State{Key: "updates", Value: s.updates},
		&pbg.State{Key: "records_in_update", Value: s.recordsInUpdate},
	}
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

	server.RegisterServer("recordprocess", false)
	server.RegisterRepeatingTask(server.processRecords, "process_records", time.Minute)
	server.Serve()
}
