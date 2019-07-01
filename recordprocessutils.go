package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordprocess/proto"
	"github.com/golang/protobuf/proto"
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

func (s *Server) processRecords(ctx context.Context) error {
	s.updates++
	startTime := time.Now()
	scoresUpdated := false
	records, err := s.getter.getRecords(ctx)

	if err != nil {
		return err
	}

	seen := make(map[int32]bool)
	for _, r := range records {
		seen[r.GetRelease().InstanceId] = true
	}

	count := int64(0)
	s.recordsInUpdate = int64(len(records))
	for _, record := range records {
		count++
		scoresUpdated = s.saveRecordScore(ctx, record) || scoresUpdated
		pre := proto.Clone(record.GetMetadata())
		update, rule := s.processRecord(record)

		if update != nil {
			s.Log(fmt.Sprintf("PRE  %v", pre))
			s.Log(fmt.Sprintf("APPL %v", rule))
			s.Log(fmt.Sprintf("POST %v", update.GetMetadata()))

			if int64(update.GetRelease().InstanceId) == s.lastUpdate {
				s.updateCount++
				if s.updateCount > 20 {
					s.RaiseIssue(ctx, "Stuck Process", fmt.Sprintf("%v is stuck in process", update.GetRelease().Id), false)
				}
			} else {
				s.updateCount = 0
			}
			s.lastUpdate = int64(update.GetRelease().InstanceId)

			s.Log(fmt.Sprintf("Updating %v and %v", update.GetRelease().Title, update.GetRelease().InstanceId))
			s.getter.update(ctx, update)
			break
		}
	}

	s.lastProc = time.Now()
	s.lastCount = count
	s.lastProcDuration = time.Now().Sub(startTime)

	if scoresUpdated {
		s.saveScores(ctx)
	}

	return nil
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

func (s *Server) processRecord(r *pbrc.Record) (*pbrc.Record, string) {
	// Don't process a record that has a pending score
	if r.GetMetadata() != nil && r.GetMetadata().SetRating != 0 {
		return nil, "Pending Score"
	}

	if r.GetMetadata() == nil {
		r.Metadata = &pbrc.ReleaseMetadata{}
	}

	if r.GetRelease().FolderId == 1782105 && r.GetMetadata().GoalFolder == 0 {
		r.GetMetadata().GoalFolder = 1782105
		return r, "Bandcamp"
	}

	if r.GetMetadata().GoalFolder == 268147 && r.GetMetadata().Category != pbrc.ReleaseMetadata_DIGITAL {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_DIGITAL
		return r, "Digital"
	}

	// If the record has no labels move it to NO_LABELS
	if len(r.GetRelease().Labels) == 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_NO_LABELS
		r.GetMetadata().Purgatory = pbrc.Purgatory_NEEDS_LABELS
		return r, "No Labels"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_NO_LABELS && len(r.GetRelease().Labels) > 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_FRESHMAN
		r.GetMetadata().Purgatory = pbrc.Purgatory_ALL_GOOD
		return r, "Found Labels"
	}

	// Deal with parents records
	if r.GetRelease().FolderId == 1727264 && r.GetMetadata().Category != pbrc.ReleaseMetadata_PARENTS {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PARENTS
		return r, "Parents"
	}

	// If the record is in google play, set the category to GOOGLE_PLAY
	if (r.GetRelease().FolderId == 1433217 || r.GetMetadata().GoalFolder == 1433217) && r.GetMetadata().Category != pbrc.ReleaseMetadata_GOOGLE_PLAY {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_GOOGLE_PLAY
		r.GetMetadata().GoalFolder = 1433217
		return r, "Google Play"
	}

	if (r.GetMetadata().Category == pbrc.ReleaseMetadata_LISTED_TO_SELL || r.GetMetadata().Category == pbrc.ReleaseMetadata_SOLD_OFFLINE) && r.GetMetadata().SaleState == pbgd.SaleState_SOLD {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_SOLD_ARCHIVE
		return r, "Sold"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_SOLD && r.GetMetadata().SalePrice > 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_SOLD_ARCHIVE
		return r, "Sold"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_SOLD && r.GetMetadata().SaleId > 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_LISTED_TO_SELL
		return r, "Listed to Sell"
	}

	if r.GetMetadata().SaleId > 0 && (r.GetMetadata().Category != pbrc.ReleaseMetadata_SOLD &&
		r.GetMetadata().Category != pbrc.ReleaseMetadata_SOLD_ARCHIVE &&
		r.GetMetadata().Category != pbrc.ReleaseMetadata_SOLD_OFFLINE &&
		r.GetMetadata().Category != pbrc.ReleaseMetadata_STALE_SALE &&
		r.GetMetadata().Category != pbrc.ReleaseMetadata_LISTED_TO_SELL &&
		r.GetMetadata().Category != pbrc.ReleaseMetadata_RIP_THEN_SELL &&
		r.GetMetadata().Category != pbrc.ReleaseMetadata_STAGED_TO_SELL &&
		r.GetMetadata().Category != pbrc.ReleaseMetadata_ASSESS_FOR_SALE &&
		r.GetMetadata().Category != pbrc.ReleaseMetadata_PREPARE_TO_SELL) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_LISTED_TO_SELL
		return r, "Listed to Sell"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_RIP_THEN_SELL && !recordNeedsRip(r) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PREPARE_TO_SELL
		return r, "Preping for sale"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_ASSESS_FOR_SALE && (r.GetMetadata().LastStockCheck > time.Now().AddDate(-1, 0, 0).Unix() || r.GetMetadata().Match == pbrc.ReleaseMetadata_FULL_MATCH) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PREPARE_TO_SELL
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PREPARE_TO_SELL {

		if r.GetMetadata().LastStockCheck < time.Now().AddDate(-1, 0, 0).Unix() && r.GetMetadata().Match != pbrc.ReleaseMetadata_FULL_MATCH {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_ASSESS_FOR_SALE
			return r, "Asessing for sale"
		}

		if recordNeedsRip(r) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_RIP_THEN_SELL
			return r, "Ripping then selling"
		}

		if r.GetRelease().Rating > 0 {
			if r.GetMetadata().SetRating == 0 {
				r.GetMetadata().SetRating = -1
				return r, "Clearing Rating"
			}
		} else {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_STAGED_TO_SELL
			return r, "Staging to Sell"
		}
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_DIGITAL && r.GetMetadata().GoalFolder != 268147 && r.GetMetadata().GoalFolder != 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_ASSESS
		return r, "Assessing"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_STAGED_TO_SELL && r.GetRelease().Rating > 0 {
		if r.GetRelease().Rating <= 3 {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_SOLD
			return r, "Sold"
		}

		if r.GetRelease().Rating == 5 {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_FRESHMAN
			return r, "Returning to fold"
		}
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PURCHASED && r.GetMetadata().Cost > 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_UNLISTENED
		return r, "New Record"
	}

	if r.GetRelease().FolderId == 1 && r.GetMetadata().Category != pbrc.ReleaseMetadata_PURCHASED && r.GetMetadata().GoalFolder <= 1 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PURCHASED
		return r, "Uncategorized record"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN {
		if r.GetMetadata().GetDateAdded() > (time.Now().AddDate(0, -3, 0).Unix()) {
			if r.GetRelease().Rating == 0 {
				r.GetMetadata().Category = pbrc.ReleaseMetadata_UNLISTENED
				return r, "Not yet listened"
			}
			r.GetMetadata().Category = pbrc.ReleaseMetadata_STAGED
			return r, "Staged"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNLISTENED {
		if r.GetRelease().Rating > 0 {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_STAGED
			return r, "Staged"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_FRESHMAN && r.GetMetadata().GetDateAdded() > time.Now().AddDate(0, -3, 0).Unix() {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_UNKNOWN
		return r, "PRE_FRESHMAN wrong"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_FRESHMAN && r.GetRelease().Rating > 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_UNKNOWN
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN {
		if r.GetMetadata().GetDateAdded() > (time.Now().AddDate(0, -6, 0).Unix()) && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -3, 0).Unix()) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_FRESHMAN
			return r, "FRESHMAN"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-3, 0, 0).Unix()) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_PROFESSOR
			return r, "PROF"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_POSTDOC && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-2, 0, 0).Unix()) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_POSTDOC
			return r, "POSTDOC"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_GRADUATE && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-1, 0, 0).Unix()) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_GRADUATE
			return r, "GRAD"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_SOPHMORE && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -6, 0).Unix()) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_SOPHMORE
			return r, "SOPHMORE"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, 2, 0).Unix()) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_HIGH_SCHOOL
			return r, "HIGH SCHOOL"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STAGED && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -2, 0).Unix()) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL
		r.GetMetadata().SetRating = -1
		return r, "PRE HS"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_HIGH_SCHOOL && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -3, 0).Unix()) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_FRESHMAN
		r.GetMetadata().SetRating = -1
		return r, "PRE F"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_FRESHMAN && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -6, 0).Unix()) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_SOPHMORE
		r.GetMetadata().SetRating = -1
		return r, "PRE S"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOPHMORE && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-1, 0, 0).Unix()) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_GRADUATE
		r.GetMetadata().SetRating = -1
		return r, "PRE G"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_GRADUATE && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-2, 0, 0).Unix()) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_POSTDOC
		r.GetMetadata().SetRating = -1
		return r, "PRE P"
	}

	return nil, "No rules applied"
}
