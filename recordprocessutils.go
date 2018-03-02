package main

import (
	"fmt"
	"strings"
	"time"

	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordprocess/proto"
)

type getter interface {
	getRecords() ([]*pbrc.Record, error)
	update(*pbrc.Record) error
}

func (s *Server) saveRecordScore(record *pbrc.Record) bool {
	found := false
	for _, score := range s.scores.GetScores() {
		if score.GetCategory() == record.GetMetadata().GetCategory() && score.GetInstanceId() == record.GetRelease().InstanceId {
			found = true
			break
		}
	}

	if !found && record.GetRelease().Rating > 0 && !strings.HasPrefix(record.GetMetadata().GetCategory().String(), "PRE") {
		s.scores.Scores = append(s.scores.Scores, &pb.RecordScore{InstanceId: record.GetRelease().InstanceId, Rating: record.GetRelease().Rating, Category: record.GetMetadata().GetCategory()})
	}

	return !found
}

func (s *Server) processRecords() {
	scoresUpdated := false
	records, err := s.getter.getRecords()

	if err != nil {
		return
	}

	count := int64(0)
	for _, record := range records {
		scoresUpdated = s.saveRecordScore(record) || scoresUpdated
		update := s.processRecord(record)
		if update != nil {
			count++
			err := s.getter.update(update)
			if err != nil {
				s.Log(fmt.Sprintf("Error updating record: %v", err))
			}
		}
	}

	s.lastProc = time.Now()
	s.lastCount = count

	if scoresUpdated {
		s.saveScores()
	}
}

func (s *Server) processRecord(r *pbrc.Record) *pbrc.Record {
	if r.GetMetadata() == nil {
		r.Metadata = &pbrc.ReleaseMetadata{}
	}

	if r.GetMetadata().Purgatory == pbrc.Purgatory_NEEDS_STOCK_CHECK && r.GetMetadata().LastStockCheck > time.Now().AddDate(0, -3, 0).Unix() {
		r.GetMetadata().Purgatory = pbrc.Purgatory_UNKNOWN
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_FRESHMAN
		return r
	}
	
	if r.GetRelease().FolderId == 1 && r.GetMetadata().Category != pbrc.ReleaseMetadata_PURCHASED {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PURCHASED
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN {
		if r.GetMetadata().GetDateAdded() > (time.Now().AddDate(0, -3, 0).Unix()) {
			if r.GetRelease().Rating == 0 {
				r.GetMetadata().Category = pbrc.ReleaseMetadata_UNLISTENED
				return r
			}
			r.GetMetadata().Category = pbrc.ReleaseMetadata_STAGED
			return r
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNLISTENED {
		if r.GetMetadata().GetDateAdded() > (time.Now().AddDate(0, -3, 0).Unix()) {
			if r.GetRelease().Rating > 0 {
				r.GetMetadata().Category = pbrc.ReleaseMetadata_STAGED
				return r
			}
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_FRESHMAN && r.GetRelease().Rating > 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_UNKNOWN
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN {
		if r.GetMetadata().GetDateAdded() > (time.Now().AddDate(0, -6, 0).Unix()) && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -3, 0).Unix()) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_FRESHMAN
			return r
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-3, 0, 0).Unix()) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_PROFESSOR
			return r
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-2, 0, 0).Unix()) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_POSTDOC
			return r
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_GRADUATE && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-1, 0, 0).Unix()) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_GRADUATE
			return r
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_SOPHMORE && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -6, 0).Unix()) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_SOPHMORE
			return r
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STAGED && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -3, 0).Unix()) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_FRESHMAN
		r.GetMetadata().SetRating = -1
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_FRESHMAN && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -6, 0).Unix()) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_SOPHMORE
		r.GetMetadata().SetRating = -1
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOPHMORE && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-1, 0, 0).Unix()) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_GRADUATE
		r.GetMetadata().SetRating = -1
		return r
	}

	return nil
}
