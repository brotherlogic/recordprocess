package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"google.golang.org/grpc"

	pbgs "github.com/brotherlogic/goserver/proto"
	pb "github.com/brotherlogic/recordprocess/proto"
	pbt "github.com/brotherlogic/tracer/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
)

func main() {
	ctx, cancel := utils.BuildContext("recordmover_cli_"+os.Args[1], "recordmover", pbgs.ContextType_MEDIUM)
	defer cancel()

	host, port, err := utils.Resolve("recordprocess")
	if err != nil {
		log.Fatalf("Unable to reach organiser: %v", err)
	}
	conn, err := grpc.Dial(host+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}

	client := pb.NewScoreServiceClient(conn)
	switch os.Args[1] {
	case "get":
		val, _ := strconv.Atoi(os.Args[2])
		res, err := client.GetScore(ctx, &pb.GetScoreRequest{InstanceId: int32(val)})
		if err != nil {
			log.Fatalf("Error on GET: %v", err)
		}
		for i, score := range res.GetScores() {
			fmt.Printf("%v. -> %v\n", i, score)
		}
	}

	utils.SendTrace(ctx, "recordmover_cli_"+os.Args[1], time.Now(), pbt.Milestone_END, "recordmover")
}
