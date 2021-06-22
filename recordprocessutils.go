package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	pbgd "github.com/brotherlogic/godiscogs"
	"github.com/brotherlogic/goserver/utils"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type getter interface {
	getRecords(ctx context.Context, t int64, c int) ([]int32, error)
	getRecord(ctx context.Context, instanceID int32) (*pbrc.Record, error)
	update(ctx context.Context, instanceID int32, category pbrc.ReleaseMetadata_Category, reason string, scount int32) error
	updateStock(ctx context.Context, rec *pbrc.Record) error
}

func (s *Server) runLoop() {
	ctx, cancel := utils.ManualContext("rp-loop", time.Minute)
	defer cancel()

	config, err := s.readConfig(ctx)
	if err != nil {
		s.Log(fmt.Sprintf("Unable to process: %v", err))
		return
	}

	bt := time.Now().Unix()
	bid := int32(-1)
	for id, t := range config.GetNextUpdateTime() {
		if t < bt {
			bt = t
			bid = id
		}

	}

	if bid > 0 {
		_, err := s.ClientUpdate(ctx, &pbrc.ClientUpdateRequest{InstanceId: bid})
		s.Log(fmt.Sprintf("Updated %v -> %v", bid, err))
	}
}

func (s *Server) isJustCd(ctx context.Context, record *pbrc.Record) bool {
	if len(record.GetRelease().GetFormats()) == 0 {
		return false
	}

	for _, format := range record.GetRelease().GetFormats() {
		if format.GetName() != "CD" && format.GetName() != "CDr" {
			return false
		}
	}

	return true
}

func (s *Server) saveRecordScore(ctx context.Context, record *pbrc.Record) bool {
	found := false
	/*for _, score := range s.scores.GetScores() {
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
	}*/

	return !found
}

var (
	backlog = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "recordprocess_backlog",
		Help: "The size of the print queue",
	})
)

func recordNeedsRip(r *pbrc.Record) bool {
	hasCD := false
	// Needs a rip if it has a CD in the formats
	for _, format := range r.GetRelease().Formats {
		if format.Name == "CD" {
			hasCD = true
		}
	}

	return hasCD && r.GetMetadata().CdPath == ""
}

func (s *Server) processRecord(ctx context.Context, r *pbrc.Record) (pbrc.ReleaseMetadata_Category, int, string) {
	// Don't process a record that has a pending score
	if r.GetMetadata() != nil && (r.GetMetadata().SetRating != 0 || r.GetMetadata().Dirty) {
		return pbrc.ReleaseMetadata_UNKNOWN, -1, "Pending Score"
	}

	if r.GetMetadata().GetGoalFolder() == 268147 && r.GetMetadata().Category != pbrc.ReleaseMetadata_DIGITAL {
		return pbrc.ReleaseMetadata_DIGITAL, -1, "Digital"
	}

	if r.GetMetadata().GetGoalFolder() == 1782105 &&
		(r.GetMetadata().Category == pbrc.ReleaseMetadata_DIGITAL ||
			r.GetMetadata().Category == pbrc.ReleaseMetadata_BANDCAMP ||
			r.GetMetadata().Category == pbrc.ReleaseMetadata_UNKNOWN) {
		return pbrc.ReleaseMetadata_UNLISTENED, -1, "BandcampOut"
	}

	// Deal with parents records
	if r.GetRelease().FolderId == 1727264 && r.GetMetadata().Category == pbrc.ReleaseMetadata_PARENTS && r.GetMetadata().GoalFolder != 1727264 {
		return pbrc.ReleaseMetadata_PRE_FRESHMAN, 1, "OutOfParents"
	}

	if r.GetRelease().FolderId == 1727264 && r.GetMetadata().Category != pbrc.ReleaseMetadata_PARENTS && (r.GetMetadata().GoalFolder == 1727264 || r.GetMetadata().GoalFolder == 0) {
		return pbrc.ReleaseMetadata_PARENTS, -1, "Parents"
	}

	// If the record is in google play, set the category to GOOGLE_PLAY
	if (r.GetRelease().GetFolderId() == 1433217 || r.GetMetadata().GetGoalFolder() == 1433217) && r.GetMetadata().GetCategory() != pbrc.ReleaseMetadata_GOOGLE_PLAY {
		return pbrc.ReleaseMetadata_GOOGLE_PLAY, -1, "Google Play"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN {
		return pbrc.ReleaseMetadata_PURCHASED, -1, "Purchased"
	}

	if (r.GetMetadata().Category == pbrc.ReleaseMetadata_LISTED_TO_SELL ||
		r.GetMetadata().Category == pbrc.ReleaseMetadata_SOLD_OFFLINE ||
		r.GetMetadata().Category == pbrc.ReleaseMetadata_STALE_SALE) && (r.GetMetadata().SaleState == pbgd.SaleState_SOLD || r.GetMetadata().GetSoldDate() > 0) {
		return pbrc.ReleaseMetadata_SOLD_ARCHIVE, -1, "Actually Sold"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_SOLD && (r.GetMetadata().SaleId > 0 || r.GetMetadata().GetSoldDate() > 0) {
		return pbrc.ReleaseMetadata_LISTED_TO_SELL, -1, "Listed to Sell"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_LISTED_TO_SELL && r.GetMetadata().GetSaleId() == 0 {
		return pbrc.ReleaseMetadata_SALE_ISSUE, -1, "Sale issue - no id"
	}

	if r.GetMetadata().SaleId < 0 && r.GetMetadata().Category == pbrc.ReleaseMetadata_LISTED_TO_SELL {
		return pbrc.ReleaseMetadata_UNLISTENED, -1, "Marking unlistened"
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
		return pbrc.ReleaseMetadata_LISTED_TO_SELL, -1, "Listed to Sell"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_RIP_THEN_SELL && !recordNeedsRip(r) {
		return pbrc.ReleaseMetadata_PREPARE_TO_SELL, -1, "Preping for sale"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_ASSESS_FOR_SALE {
		return pbrc.ReleaseMetadata_PREPARE_TO_SELL, -1, "ASSESSED_PREP_FOR_SALE"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PREPARE_TO_SELL {

		if recordNeedsRip(r) {
			return pbrc.ReleaseMetadata_RIP_THEN_SELL, -1, "Ripping then selling"
		}

		if r.GetRelease().Rating <= 0 {
			return pbrc.ReleaseMetadata_STAGED_TO_SELL, -1, "Staging to Sell"
		}
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_DIGITAL && r.GetMetadata().GoalFolder != 268147 && r.GetMetadata().GoalFolder != 0 {
		return pbrc.ReleaseMetadata_ASSESS, -1, "Assessing"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_STAGED_TO_SELL && r.GetRelease().Rating > 0 {
		if r.GetRelease().Rating <= 3 {
			return pbrc.ReleaseMetadata_SOLD, -1, "Sold"
		}

		if r.GetRelease().Rating == 5 {
			return pbrc.ReleaseMetadata_IN_COLLECTION, -1, "Returning to fold"
		}
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_SOPHMORE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_GRADUATE ||
		r.GetMetadata().Category == pbrc.ReleaseMetadata_FRESHMAN ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_POSTDOC ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PROFESSOR ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_DISTINGUISHED {
		return pbrc.ReleaseMetadata_IN_COLLECTION, -1, "MoveToIn"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PRE_SOPHMORE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_GRADUATE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_POSTDOC ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_PROFESSOR ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_DISTINGUISHED {
		return pbrc.ReleaseMetadata_PRE_IN_COLLECTION, -1, "PRE TO IN"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PRE_IN_COLLECTION && r.GetRelease().Rating > 0 {
		return pbrc.ReleaseMetadata_IN_COLLECTION, -1, "PlaceInCollection"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PRE_VALIDATE && r.GetRelease().Rating > 0 {
		return pbrc.ReleaseMetadata_VALIDATE, -1, "Validated"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PURCHASED && r.GetMetadata().Cost > 0 {
		return pbrc.ReleaseMetadata_UNLISTENED, -1, "New Record"
	}

	if r.GetRelease().FolderId == 1 && r.GetMetadata().Category != pbrc.ReleaseMetadata_PURCHASED && r.GetMetadata().GoalFolder <= 1 {
		return pbrc.ReleaseMetadata_PURCHASED, -1, "Uncategorized record"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNLISTENED {
		if r.GetRelease().Rating > 0 {
			return pbrc.ReleaseMetadata_STAGED, 1, "Staged"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_FRESHMAN && r.GetMetadata().GetDateAdded() > time.Now().AddDate(0, -3, 0).Unix() {
		return pbrc.ReleaseMetadata_UNLISTENED, -1, "PRE_FRESHMAN wrong"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_FRESHMAN && r.GetRelease().Rating > 0 {
		if r.GetMetadata().GetDateAdded() > (time.Now().AddDate(0, -6, 0).Unix()) && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -3, 0).Unix()) {
			return pbrc.ReleaseMetadata_IN_COLLECTION, 3, "FRESHMAN"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_DISTINGUISHED && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-4, 0, 0).Unix()) {
			return pbrc.ReleaseMetadata_DISTINGUISHED, 3, "DIST"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_PROFESSOR && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-3, 0, 0).Unix()) {
			return pbrc.ReleaseMetadata_PROFESSOR, 3, "PROF"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_POSTDOC && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-2, 0, 0).Unix()) {
			return pbrc.ReleaseMetadata_POSTDOC, 3, "POSTDOC"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_GRADUATE && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-1, 0, 0).Unix()) {
			return pbrc.ReleaseMetadata_IN_COLLECTION, 3, "GRAD"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_SOPHMORE && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -6, 0).Unix()) {
			return pbrc.ReleaseMetadata_SOPHMORE, 3, "SOPHMORE"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_FRESHMAN && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -3, 0).Unix()) {
			return pbrc.ReleaseMetadata_IN_COLLECTION, 3, "FRESHMAN"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -1, 0).Unix()) {
			return pbrc.ReleaseMetadata_HIGH_SCHOOL, 1, "HIGH SCHOOL"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STAGED && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -1, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL, -1, "PRE HS"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_HIGH_SCHOOL && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -3, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_IN_COLLECTION, -1, "PRE IN"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_FRESHMAN && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -6, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_IN_COLLECTION, -1, "PRE S"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOPHMORE && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-1, 0, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_IN_COLLECTION, -1, "PRE G"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_GRADUATE && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-2, 0, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_POSTDOC, -1, "PRE P"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_POSTDOC && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-3, 0, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_PROFESSOR, -1, "PRE PREOG"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PROFESSOR && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-4, 0, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_DISTINGUISHED, -1, "PRE DISTIN"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_ASSESS {
		return pbrc.ReleaseMetadata_PRE_FRESHMAN, -1, "ASSESSED"
	}

	return pbrc.ReleaseMetadata_UNKNOWN, -1, "No rules applied"
}
