package main

import (
	"context"
	"testing"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordprocess/proto"
)

func TestGetScore(t *testing.T) {
	s := InitTest()

	scores, err := s.GetScore(context.Background(), &pb.GetScoreRequest{InstanceId: 1234})

	if err != nil {
		t.Fatalf("Error in getting score: %v", err)
	}

	if len(scores.GetScores()) != 1 {
		t.Errorf("Bad scores: %v", scores)
	}
}

func TestForceUpdate(t *testing.T) {
	s := InitTest()
	s.getter = &testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 1727264, Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PARENTS, GoalFolder: 242017, Cost: 12}}}
	_, err := s.Force(context.Background(), &pb.ForceRequest{InstanceId: 1234})

	if err != nil {
		t.Errorf("Error in proc: %v", err)
	}
}

func TestForceUpdateFailGet(t *testing.T) {
	s := InitTest()
	s.getter = &testGetter{getFail: true}
	_, err := s.Force(context.Background(), &pb.ForceRequest{InstanceId: 1234})

	if err == nil {
		t.Errorf("No fail on get")
	}
}

func TestForceUpdateFailProcess(t *testing.T) {
	s := InitTest()
	s.getter = &testGetter{rec: &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PURCHASED}, Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1}}}
	_, err := s.Force(context.Background(), &pb.ForceRequest{InstanceId: 1234})

	if err == nil {
		t.Errorf("No fail on process")
	}
}
