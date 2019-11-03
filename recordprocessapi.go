package main

import (
	"fmt"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/recordprocess/proto"
)

// GetScore gets the score for an instance
func (s *Server) GetScore(ctx context.Context, req *pb.GetScoreRequest) (*pb.GetScoreResponse, error) {
	response := &pb.GetScoreResponse{Scores: []*pb.RecordScore{}}

	for _, score := range s.scores.GetScores() {
		if score.GetInstanceId() == req.GetInstanceId() {
			response.Scores = append(response.Scores, score)
		}
	}

	return response, nil
}

// Force a process of a specific record
func (s *Server) Force(ctx context.Context, req *pb.ForceRequest) (*pb.ForceResponse, error) {
	record, err := s.getter.getRecord(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	update, result := s.processRecord(record)
	if update != nil {
		err := s.getter.update(ctx, update)
		return &pb.ForceResponse{Result: update}, err
	}

	return nil, fmt.Errorf("Unable to process: %v", result)
}
