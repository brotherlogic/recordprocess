package main

import (
	"strings"
	"time"

	"golang.org/x/net/context"

	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordprocess/proto"
)

type getter interface {
	getRecords(ctx context.Context) ([]*pbrc.Record, error)
	update(ctx context.Context, r *pbrc.Record) error
	moveToSold(ctx context.Context, r *pbrc.Record)
}

func (s *Server) saveRecordScore(ctx context.Context, record *pbrc.Record) bool {
	found := false
	for _, score := range s.scores.GetScores() {
		if score.GetCategory() == record.GetMetadata().GetCategory() && score.GetInstanceId() == record.GetRelease().InstanceId {
			found = true
			break
		}
	}

	if !found && record.GetRelease().Rating > 0 && !strings.HasPrefix(record.GetMetadata().GetCategory().String(), "PRE") {
		s.scores.Scores = append(s.scores.Scores, &pb.RecordScore{
			InstanceId: record.GetRelease().InstanceId,
			Rating:     record.GetRelease().Rating,
			Category:   record.GetMetadata().GetCategory(),
			ScoreTime:  time.Now().Unix(),
		})
	}

	return !found
}

func (s *Server) processRecords(ctx context.Context) {
	scoresUpdated := false
	records, err := s.getter.getRecords(ctx)

	if err != nil {
		return
	}

	count := int64(0)
	for _, record := range records {
		scoresUpdated = s.saveRecordScore(ctx, record) || scoresUpdated
		update := s.processRecord(record)
		if update != nil {
			count++
			s.getter.update(ctx, update)
		}
	}

	s.lastProc = time.Now()
	s.lastCount = count

	if scoresUpdated {
		s.saveScores(ctx)
	}
}

func recordNeedsRip(r *pbrc.Record) bool {
	hasCD := false
	// Needs a rip if it has a CD in the formats
	for _, format := range r.GetRelease().Formats {
		if format.Name == "CD" {
			hasCD = true
		}
	}

	return hasCD && r.GetMetadata().FilePath == ""
}

func (s *Server) processRecord(r *pbrc.Record) *pbrc.Record {
	// Don't process a record that has a pending score
	if r.GetMetadata() != nil && r.GetMetadata().SetRating != 0 {
		return nil
	}

	if r.GetMetadata() == nil {
		r.Metadata = &pbrc.ReleaseMetadata{}
	}

	// If the record has no labels move it to NO_LABELS
	if len(r.GetRelease().Labels) == 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_NO_LABELS
		r.GetMetadata().Purgatory = pbrc.Purgatory_NEEDS_LABELS
		return r
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_NO_LABELS && len(r.GetRelease().Labels) > 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_FRESHMAN
		r.GetMetadata().Purgatory = pbrc.Purgatory_ALL_GOOD
		return r
	}

	// If the record is in google play, set the category to GOOGLE_PLAY
	if r.GetRelease().FolderId == 1433217 && r.GetMetadata().Category != pbrc.ReleaseMetadata_GOOGLE_PLAY {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_GOOGLE_PLAY
		return r
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_SOLD && r.GetMetadata().SalePrice > 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_SOLD_ARCHIVE
		return r
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_SOLD && r.GetMetadata().SaleId > 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_LISTED_TO_SELL
		return r
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_RIP_THEN_SELL && !recordNeedsRip(r) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PREPARE_TO_SELL
		return r
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PREPARE_TO_SELL {

		if time.Now().Sub(time.Unix(r.GetMetadata().LastStockCheck, 0)) > time.Hour*24*30 {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_ASSESS_FOR_SALE
			return r
		}

		if recordNeedsRip(r) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_RIP_THEN_SELL
			return r
		}

		if r.GetRelease().Rating > 0 {
			if r.GetMetadata().SetRating == 0 {
				r.GetMetadata().SetRating = -1
				return r
			}
		} else {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_STAGED_TO_SELL
			return r
		}
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_DIGITAL && r.GetMetadata().GoalFolder != 268147 && r.GetMetadata().GoalFolder != 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_ASSESS
		return r
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_STAGED_TO_SELL && r.GetRelease().Rating > 0 {
		if r.GetRelease().Rating <= 3 {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_SOLD
			return r
		}

		if r.GetRelease().Rating == 5 {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_FRESHMAN
			return r
		}
	}

	if r.GetMetadata().Purgatory == pbrc.Purgatory_NEEDS_STOCK_CHECK && r.GetMetadata().LastStockCheck > time.Now().AddDate(0, -3, 0).Unix() {
		r.GetMetadata().Purgatory = pbrc.Purgatory_ALL_GOOD
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

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_POSTDOC && r.GetRelease().Rating > 0) {
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

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_GRADUATE && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-2, 0, 0).Unix()) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_POSTDOC
		r.GetMetadata().SetRating = -1
		return r
	}

	return nil
}
