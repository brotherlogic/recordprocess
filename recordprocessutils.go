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
	getRecords(ctx context.Context, t int64) ([]int32, error)
	getRecord(ctx context.Context, instanceID int32) (*pbrc.Record, error)
	update(ctx context.Context, r *pbrc.Record) error
	moveToSold(ctx context.Context, r *pbrc.Record)
}

func (s *Server) isJustCd(ctx context.Context, record *pbrc.Record) bool {
	if len(record.GetRelease().GetFormats()) == 0 {
		return false
	}

	for _, format := range record.GetRelease().GetFormats() {
		if format.GetName() != "CD" {
			return false
		}
	}

	return true
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
	records, err := s.getter.getRecords(ctx, s.config.LastRunTime)
	if err != nil {
		return err
	}

	if len(records) > 100 {
		s.RaiseIssue(ctx, "Errr", fmt.Sprintf("Big addition to next update time[from %v]: %v", s.config.LastRunTime, len(records)), false)
	}
	s.configMutex.Lock()
	for _, instanceID := range records {
		s.config.NextUpdateTime[instanceID] = time.Now().Unix()
	}
	s.config.LastRunTime = time.Now().Unix()
	s.configMutex.Unlock()

	return s.saveConfig(ctx)
}

func (s *Server) processNextRecords(ctx context.Context) error {
	for instanceID, timev := range s.config.NextUpdateTime {
		if time.Unix(timev, 0).Before(time.Now()) {
			record, err := s.getter.getRecord(ctx, instanceID)
			if err != nil {
				return err
			}
			scoresUpdated := s.saveRecordScore(ctx, record)
			pre := proto.Clone(record.GetMetadata())
			update, rule := s.processRecord(ctx, record)

			if update != nil {
				s.Log(fmt.Sprintf("APPL %v -> %v -> %v", pre, rule, update.GetMetadata()))

				if int64(update.GetRelease().InstanceId) == s.lastUpdate {
					s.updateCount++
					if s.updateCount > 20 {
						s.RaiseIssue(ctx, "Stuck Process", fmt.Sprintf("%v is stuck in process [Last rule applied: %v]", update.GetRelease().Id, rule), false)
					}
				} else {
					s.updateCount = 0
				}
				s.lastUpdate = int64(update.GetRelease().InstanceId)

				s.Log(fmt.Sprintf("Updating %v and %v", update.GetRelease().Title, update.GetRelease().InstanceId))
				err := s.getter.update(ctx, update)
				s.Log(fmt.Sprintf("FAILURE TO UPDATE: %v", err))
			}

			s.lastProc = time.Now()
			s.configMutex.Lock()
			delete(s.config.NextUpdateTime, instanceID)
			s.configMutex.Unlock()

			if scoresUpdated {
				s.saveScores(ctx)
			}
			return s.saveConfig(ctx)
		}
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

func (s *Server) processRecord(ctx context.Context, r *pbrc.Record) (*pbrc.Record, string) {
	s.Log(fmt.Sprintf("Processing %v -> %v", r.GetRelease().GetInstanceId(), r.GetRelease().GetTitle()))

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
	if r.GetRelease().FolderId == 1727264 && r.GetMetadata().Category == pbrc.ReleaseMetadata_PARENTS && r.GetMetadata().GoalFolder != 1727264 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_FRESHMAN
		return r, "OutOfParents"
	}

	if r.GetRelease().FolderId == 1727264 && r.GetMetadata().Category != pbrc.ReleaseMetadata_PARENTS && (r.GetMetadata().GoalFolder == 1727264 || r.GetMetadata().GoalFolder == 0) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PARENTS
		return r, "Parents"
	}

	// If the record is in google play, set the category to GOOGLE_PLAY
	if (r.GetRelease().FolderId == 1433217 || r.GetMetadata().GoalFolder == 1433217) && r.GetMetadata().Category != pbrc.ReleaseMetadata_GOOGLE_PLAY {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_GOOGLE_PLAY
		r.GetMetadata().GoalFolder = 1433217
		return r, "Google Play"
	}

	if (r.GetMetadata().Category == pbrc.ReleaseMetadata_LISTED_TO_SELL ||
		r.GetMetadata().Category == pbrc.ReleaseMetadata_SOLD_OFFLINE ||
		r.GetMetadata().Category == pbrc.ReleaseMetadata_STALE_SALE) && r.GetMetadata().SaleState == pbgd.SaleState_SOLD {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_SOLD_ARCHIVE
		return r, "Sold"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_SOLD && r.GetMetadata().SaleId > 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_LISTED_TO_SELL
		return r, "Listed to Sell"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_LISTED_TO_SELL && r.GetMetadata().GetSaleId() == 0 {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_SALE_ISSUE
		return r, "Sale issue - no id"
	}

	if r.GetMetadata().SaleId < 0 && r.GetMetadata().Category == pbrc.ReleaseMetadata_LISTED_TO_SELL {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_UNLISTENED
		return r, "Marking unlistened"
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

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_ASSESS_FOR_SALE && (r.GetMetadata().LastStockCheck > time.Now().AddDate(-1, 0, 0).Unix() || r.GetMetadata().Match == pbrc.ReleaseMetadata_FULL_MATCH || s.isJustCd(ctx, r)) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PREPARE_TO_SELL
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PREPARE_TO_SELL {

		if (r.GetMetadata().LastStockCheck < time.Now().AddDate(-1, 0, 0).Unix() && r.GetMetadata().Match != pbrc.ReleaseMetadata_FULL_MATCH && !s.isJustCd(ctx, r)) && r.GetMetadata().Match != pbrc.ReleaseMetadata_FULL_MATCH {
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

		r.GetMetadata().SetRating = -1
		return r, "Clearing Rating when Staged"
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
		r.GetMetadata().Category = pbrc.ReleaseMetadata_UNLISTENED
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

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_DISTINGUISHED && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-4, 0, 0).Unix()) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_DISTINGUISHED
			return r, "DIST"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_PROFESSOR && r.GetRelease().Rating > 0) {
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
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -1, 0).Unix()) {
			r.GetMetadata().Category = pbrc.ReleaseMetadata_HIGH_SCHOOL
			return r, "HIGH SCHOOL"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STAGED && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -1, 0).Unix()) {
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

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_POSTDOC && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-3, 0, 0).Unix()) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_PROFESSOR
		r.GetMetadata().SetRating = -1
		return r, "PRE PREOG"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PROFESSOR && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-4, 0, 0).Unix()) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_DISTINGUISHED
		r.GetMetadata().SetRating = -1
		return r, "PRE DISTIN"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_ASSESS && r.GetMetadata().GetPurgatory() == pbrc.Purgatory_NEEDS_STOCK_CHECK &&
		(time.Now().Sub(time.Unix(r.GetMetadata().GetLastStockCheck(), 0)) < time.Hour*24*7*4) {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_FRESHMAN
		r.GetMetadata().Purgatory = -1
		return r, "ASSESSED"
	}

	return nil, "No rules applied"
}
