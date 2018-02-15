package main

import (
	"context"
	"testing"

	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordprocess/proto"
)

func TestGetScore(t *testing.T) {
	s := InitTest()

	s.scores.Scores = append(s.scores.Scores, &pb.RecordScore{
		InstanceId: 1234,
		Rating:     5,
		Category:   pbrc.ReleaseMetadata_FRESHMAN,
	})

	scores, err := s.GetScore(context.Background(), &pb.GetScoreRequest{InstanceId: 1234})

	if err != nil {
		t.Fatalf("Error in getting score: %v", err)
	}

	if len(scores.GetScores()) != 1 {
		t.Errorf("Bad scores: %v", scores)
	}
}
