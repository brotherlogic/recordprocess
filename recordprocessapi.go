package main

import (
	"fmt"

	"golang.org/x/net/context"

	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordprocess/proto"
)

// GetScore gets the score for an instance
func (s *Server) GetScore(ctx context.Context, req *pb.GetScoreRequest) (*pb.GetScoreResponse, error) {
	response := &pb.GetScoreResponse{Scores: []*pb.RecordScore{}}

	for _, score := range s.scores.GetScores() {
		if score.GetInstanceId() == req.GetInstanceId() && score.GetRating() > 0 {
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

	update, result := s.processRecord(ctx, record)
	if update != pbrc.ReleaseMetadata_UNKNOWN {
		err := s.getter.update(ctx, record.GetRelease().GetInstanceId(), update, result)
		return &pb.ForceResponse{Result: update, Reason: result}, err
	}

	return nil, fmt.Errorf("Unable to process: %v", result)
}

//ClientUpdate forces a move
func (s *Server) ClientUpdate(ctx context.Context, in *pbrc.ClientUpdateRequest) (*pbrc.ClientUpdateResponse, error) {
	record, err := s.getter.getRecord(ctx, in.InstanceId)
	if err != nil {
		return nil, err
	}

	update, result := s.processRecord(ctx, record)
	if update != pbrc.ReleaseMetadata_UNKNOWN {
		err := s.getter.update(ctx, record.GetRelease().GetInstanceId(), update, result)
		return &pbrc.ClientUpdateResponse{}, err
	}

	return nil, fmt.Errorf("Unable to process: %v", result)
}
