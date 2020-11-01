package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver/utils"
	pbrc "github.com/brotherlogic/recordcollection/proto"
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
		res, err := client.Get(ctx, &pb.GetRequest{InstanceId: int32(val)})
		if err != nil {
			log.Fatalf("Error on GET: %v", err)
		}
		fmt.Printf("%v -> %v\n", os.Args[2], time.Unix(res.GetNextUpdateTime(), 0))
	case "force":
		client2 := pbrc.NewClientUpdateServiceClient(conn)
		val, _ := strconv.Atoi(os.Args[2])
		res, err := client2.ClientUpdate(ctx, &pbrc.ClientUpdateRequest{InstanceId: int32(val)})
		fmt.Printf("%v\n%v", res, err)

	}
}
