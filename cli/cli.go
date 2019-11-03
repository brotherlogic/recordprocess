package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/recordprocess/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
)

func main() {
	ctx, cancel := utils.BuildContext("recordmover_cli_"+os.Args[1], "recordmover")
	defer cancel()

	host, port, err := utils.Resolve("recordprocess", "recordprocess-cli")
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
			fmt.Printf("%v. [%v] -> %v,%v\n", i, time.Unix(score.ScoreTime, 0), score.Rating, score.Category)
		}
	case "force":
		val, _ := strconv.Atoi(os.Args[2])
		res, err := client.Force(ctx, &pb.ForceRequest{InstanceId: int32(val)})
		fmt.Printf("%v\n%v", res, err)

	}
}
