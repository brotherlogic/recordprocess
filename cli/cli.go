package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver/utils"
	pb "github.com/brotherlogic/recordprocess/proto"

	//Needed to pull in gzip encoding init

	_ "google.golang.org/grpc/encoding/gzip"
)

func main() {
	ctx, cancel := utils.ManualContext("recordprocess-cli", "recordprocess-cli", time.Minute, true)
	defer cancel()

	conn, err := utils.LFDialServer(ctx, "recordprocess")
	if err != nil {
		log.Fatalf("Can't dial getter: %v", err)
	}

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
