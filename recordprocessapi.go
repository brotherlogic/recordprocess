package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordprocess/proto"
)

// GetScore gets the score for an instance
func (s *Server) GetScore(ctx context.Context, req *pb.GetScoreRequest) (*pb.GetScoreResponse, error) {
	scores, err := s.readScores(ctx)
	if err != nil {
		return nil, err
	}

	response := &pb.GetScoreResponse{Scores: []*pb.RecordScore{}}

	for _, score := range scores.GetScores() {
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

	update, _, result := s.processRecord(ctx, record)
	if update != pbrc.ReleaseMetadata_UNKNOWN {
		ncount := record.GetMetadata().GetSaleAttempts()
		if update == pbrc.ReleaseMetadata_LISTED_TO_SELL {
			ncount++
		}

		err := s.getter.update(ctx, record.GetRelease().GetInstanceId(), update, result, ncount)
		return &pb.ForceResponse{Result: update, Reason: result}, err
	}

	return nil, fmt.Errorf("Unable to process: %v", result)
}

//ClientUpdate forces a move
func (s *Server) ClientUpdate(ctx context.Context, in *pbrc.ClientUpdateRequest) (*pbrc.ClientUpdateResponse, error) {
	record, err := s.getter.getRecord(ctx, in.InstanceId)
	if err != nil {
		code := status.Convert(err).Code()

		// This record has been deleted - remove it from processing
		if code == codes.OutOfRange {
			config, err := s.readConfig(ctx)
			if err == nil {
				delete(config.NextUpdateTime, in.InstanceId)
				return &pbrc.ClientUpdateResponse{}, s.saveConfig(ctx, config)
			}

		}
		return nil, err
	}

	update, _, result := s.processRecord(ctx, record)
	if update != pbrc.ReleaseMetadata_UNKNOWN {
		ncount := record.GetMetadata().GetSaleAttempts()
		if update == pbrc.ReleaseMetadata_STAGED_TO_SELL {
			ncount++
		}
		err := s.getter.update(ctx, record.GetRelease().GetInstanceId(), update, result, ncount)
		s.Log(fmt.Sprintf("%v -> %v, %v => %v", record.GetRelease().GetTitle(), update, result, err))
		if err != nil {
			return nil, err
		}
	}

	/*if ti >= 0 {
		return &pbrc.ClientUpdateResponse{}, s.updateTime(ctx, in.InstanceId, time.Now().Add(time.Duration(ti)*time.Hour*24*7*30).Unix())
	}*/

	return &pbrc.ClientUpdateResponse{}, s.updateTime(ctx, in.InstanceId, time.Now().Add(time.Hour*24*7).Unix())
}

//Get peek into the state
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	config, err := s.readConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetResponse{NextUpdateTime: config.GetNextUpdateTime()[req.GetInstanceId()]}, nil
}
