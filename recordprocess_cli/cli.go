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
	ctx, cancel := utils.ManualContext("recordprocess-cli", time.Minute)
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
	case "fullping":
		ctx2, cancel2 := utils.ManualContext("recordcollectioncli-"+os.Args[1], time.Hour)
		defer cancel2()

		conn2, err := utils.LFDialServer(ctx2, "recordcollection")
		if err != nil {
			log.Fatalf("Cannot reach rc: %v", err)
		}
		defer conn2.Close()

		registry := pbrc.NewRecordCollectionServiceClient(conn2)
		ids, err := registry.QueryRecords(ctx2, &pbrc.QueryRecordsRequest{Query: &pbrc.QueryRecordsRequest_FolderId{673768}})
		if err != nil {
			log.Fatalf("Bad query: %v", err)
		}
		client2 := pbrc.NewClientUpdateServiceClient(conn)
		for i, id := range ids.GetInstanceIds() {
			log.Printf("PING %v -> %v", i, id)
			ctx3, cancel3 := utils.ManualContext("fullping", time.Minute)
			res, err := client2.ClientUpdate(ctx3, &pbrc.ClientUpdateRequest{InstanceId: int32(id)})
			fmt.Printf("%v\n%v", res, err)
			cancel3()
		}

		ids, err = registry.QueryRecords(ctx2, &pbrc.QueryRecordsRequest{Query: &pbrc.QueryRecordsRequest_FolderId{3578980}})
		if err != nil {
			log.Fatalf("Bad query: %v", err)
		}
		for i, id := range ids.GetInstanceIds() {
			log.Printf("PING %v -> %v", i, id)
			ctx3, cancel3 := utils.ManualContext("fullping", time.Minute)
			res, err := client2.ClientUpdate(ctx3, &pbrc.ClientUpdateRequest{InstanceId: int32(id)})
			fmt.Printf("%v\n%v", res, err)
			cancel3()
		}
	}
}
