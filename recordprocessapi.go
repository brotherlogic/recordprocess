package main

import "golang.org/x/net/context"

import pb "github.com/brotherlogic/recordprocess/proto"

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
