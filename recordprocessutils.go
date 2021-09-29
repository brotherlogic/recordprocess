package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	pbgd "github.com/brotherlogic/godiscogs"
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

func (s *Server) runLoop(ctx context.Context) {

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

var (
	NO_CHANGE = time.Unix(0, 0)
)

func (s *Server) processRecord(ctx context.Context, r *pbrc.Record) (pbrc.ReleaseMetadata_Category, time.Time, string) {
	// Don't process a record that has a pending score
	if r.GetMetadata() != nil && (r.GetMetadata().SetRating != 0 || r.GetMetadata().Dirty) {
		return pbrc.ReleaseMetadata_UNKNOWN, NO_CHANGE, "Pending Score"
	}

	// Don't process a record that is in the box
	if r.GetMetadata() != nil &&
		r.GetMetadata().GetBoxState() != pbrc.ReleaseMetadata_BOX_UNKNOWN &&
		r.GetMetadata().GetBoxState() != pbrc.ReleaseMetadata_OUT_OF_BOX {
		return pbrc.ReleaseMetadata_UNKNOWN, NO_CHANGE, "In The Box"
	}

	if r.GetMetadata().GetGoalFolder() == 268147 && r.GetMetadata().Category != pbrc.ReleaseMetadata_DIGITAL {
		return pbrc.ReleaseMetadata_DIGITAL, NO_CHANGE, "Digital"
	}

	// Deal with parents records
	if r.GetRelease().FolderId == 1727264 && r.GetMetadata().Category == pbrc.ReleaseMetadata_PARENTS && r.GetMetadata().GoalFolder != 1727264 {
		return pbrc.ReleaseMetadata_PRE_FRESHMAN, NO_CHANGE, "OutOfParents"
	}

	if r.GetRelease().FolderId == 1727264 && r.GetMetadata().Category != pbrc.ReleaseMetadata_PARENTS && (r.GetMetadata().GoalFolder == 1727264 || r.GetMetadata().GoalFolder == 0) {
		return pbrc.ReleaseMetadata_PARENTS, NO_CHANGE, "Parents"
	}

	// If the record is in google play, set the category to GOOGLE_PLAY
	if (r.GetRelease().GetFolderId() == 1433217 || r.GetMetadata().GetGoalFolder() == 1433217) && r.GetMetadata().GetCategory() != pbrc.ReleaseMetadata_GOOGLE_PLAY {
		return pbrc.ReleaseMetadata_GOOGLE_PLAY, NO_CHANGE, "Google Play"
	}

	if (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PURCHASED) && r.GetMetadata().GetDateArrived() > 0 {
		return pbrc.ReleaseMetadata_ARRIVED, NO_CHANGE, "Purchased"
	}

	if (r.GetMetadata().Category == pbrc.ReleaseMetadata_LISTED_TO_SELL ||
		r.GetMetadata().Category == pbrc.ReleaseMetadata_SOLD_OFFLINE ||
		r.GetMetadata().Category == pbrc.ReleaseMetadata_STALE_SALE) && (r.GetMetadata().SaleState == pbgd.SaleState_SOLD || r.GetMetadata().GetSoldDate() > 0) {
		return pbrc.ReleaseMetadata_SOLD_ARCHIVE, NO_CHANGE, "Actually Sold"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_SOLD && (r.GetMetadata().SaleId > 0 || r.GetMetadata().GetSoldDate() > 0) {
		return pbrc.ReleaseMetadata_LISTED_TO_SELL, NO_CHANGE, "Listed to Sell"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_LISTED_TO_SELL && r.GetMetadata().GetSaleId() == 0 {
		return pbrc.ReleaseMetadata_SALE_ISSUE, NO_CHANGE, "Sale issue - no id"
	}

	if r.GetMetadata().SaleId < 0 && r.GetMetadata().Category == pbrc.ReleaseMetadata_LISTED_TO_SELL {
		return pbrc.ReleaseMetadata_UNLISTENED, NO_CHANGE, "Marking unlistened"
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
		return pbrc.ReleaseMetadata_LISTED_TO_SELL, NO_CHANGE, "Listed to Sell"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_RIP_THEN_SELL && !recordNeedsRip(r) {
		return pbrc.ReleaseMetadata_PREPARE_TO_SELL, NO_CHANGE, "Preping for sale"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_ASSESS_FOR_SALE {
		return pbrc.ReleaseMetadata_PREPARE_TO_SELL, NO_CHANGE, "ASSESSED_PREP_FOR_SALE"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PREPARE_TO_SELL {

		if recordNeedsRip(r) {
			return pbrc.ReleaseMetadata_RIP_THEN_SELL, NO_CHANGE, "Ripping then selling"
		}

		if r.GetRelease().Rating <= 0 {
			return pbrc.ReleaseMetadata_STAGED_TO_SELL, NO_CHANGE, "Staging to Sell"
		}
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_DIGITAL && r.GetMetadata().GoalFolder != 268147 && r.GetMetadata().GoalFolder != 0 {
		return pbrc.ReleaseMetadata_ASSESS, NO_CHANGE, "Assessing"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_STAGED_TO_SELL && r.GetRelease().Rating > 0 {
		if r.GetRelease().Rating <= 3 {
			return pbrc.ReleaseMetadata_SOLD, NO_CHANGE, "Sold"
		}

		if r.GetRelease().Rating == 5 {
			return pbrc.ReleaseMetadata_IN_COLLECTION, NO_CHANGE, "Returning to fold"
		}
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_SOPHMORE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_GRADUATE ||
		r.GetMetadata().Category == pbrc.ReleaseMetadata_FRESHMAN ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_POSTDOC ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PROFESSOR ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_DISTINGUISHED {
		return pbrc.ReleaseMetadata_IN_COLLECTION, NO_CHANGE, "MoveToIn"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PRE_SOPHMORE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_GRADUATE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_POSTDOC ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_PROFESSOR ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_DISTINGUISHED {
		return pbrc.ReleaseMetadata_PRE_IN_COLLECTION, NO_CHANGE, "PRE TO IN"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PRE_IN_COLLECTION && r.GetRelease().Rating > 0 {
		return pbrc.ReleaseMetadata_IN_COLLECTION, NO_CHANGE, "PlaceInCollection"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PRE_VALIDATE && r.GetRelease().Rating > 0 {
		return pbrc.ReleaseMetadata_VALIDATE, NO_CHANGE, "Validated"
	}

	if r.GetMetadata().Category == pbrc.ReleaseMetadata_PURCHASED && r.GetMetadata().Cost > 0 {
		return pbrc.ReleaseMetadata_UNLISTENED, NO_CHANGE, "New Record"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNLISTENED {
		if r.GetRelease().Rating > 0 {
			return pbrc.ReleaseMetadata_STAGED, time.Unix(r.GetMetadata().GetLastListenTime(), 0).Add(time.Hour * 24 * 30), "Staged"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_FRESHMAN && r.GetMetadata().GetDateAdded() > time.Now().AddDate(0, -3, 0).Unix() {
		return pbrc.ReleaseMetadata_UNLISTENED, NO_CHANGE, "PRE_FRESHMAN wrong"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_FRESHMAN && r.GetRelease().Rating > 0 {
		if r.GetMetadata().GetDateAdded() > (time.Now().AddDate(0, -6, 0).Unix()) && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -3, 0).Unix()) {
			return pbrc.ReleaseMetadata_IN_COLLECTION, NO_CHANGE, "FRESHMAN"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_DISTINGUISHED && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-4, 0, 0).Unix()) {
			return pbrc.ReleaseMetadata_DISTINGUISHED, NO_CHANGE, "DIST"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_PROFESSOR && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-3, 0, 0).Unix()) {
			return pbrc.ReleaseMetadata_PROFESSOR, NO_CHANGE, "PROF"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_POSTDOC && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-2, 0, 0).Unix()) {
			return pbrc.ReleaseMetadata_POSTDOC, NO_CHANGE, "POSTDOC"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_GRADUATE && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-1, 0, 0).Unix()) {
			return pbrc.ReleaseMetadata_IN_COLLECTION, NO_CHANGE, "GRAD"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_SOPHMORE && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -6, 0).Unix()) {
			return pbrc.ReleaseMetadata_SOPHMORE, NO_CHANGE, "SOPHMORE"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_FRESHMAN && r.GetRelease().Rating > 0) {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -3, 0).Unix()) {
			return pbrc.ReleaseMetadata_IN_COLLECTION, NO_CHANGE, "FRESHMAN"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL && r.GetRelease().Rating > 0) {
		return pbrc.ReleaseMetadata_HIGH_SCHOOL, time.Unix(r.GetMetadata().GetLastListenTime(), 0).Add(time.Hour * 24 * 60), "HIGH SCHOOL W/ARR"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_VALIDATE && r.GetRelease().Rating > 0 {
		return pbrc.ReleaseMetadata_STAGED, NO_CHANGE, "STAGED"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STAGED && r.GetMetadata().GetLastListenTime() < (time.Now().AddDate(0, -1, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL, NO_CHANGE, "PRE HS"
	}

	s.Log(fmt.Sprintf("%v -> %v and %v", r.GetRelease().GetInstanceId(), time.Unix(r.GetMetadata().GetDateAdded(), 0), time.Now().AddDate(0, -3, 0)))
	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_HIGH_SCHOOL && r.GetMetadata().GetLastListenTime() < (time.Now().AddDate(0, -2, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_IN_COLLECTION, NO_CHANGE, "PRE IN"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_FRESHMAN && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -6, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_IN_COLLECTION, NO_CHANGE, "PRE S"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOPHMORE && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-1, 0, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_IN_COLLECTION, NO_CHANGE, "PRE G"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_GRADUATE && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-2, 0, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_POSTDOC, NO_CHANGE, "PRE P"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_POSTDOC && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-3, 0, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_PROFESSOR, NO_CHANGE, "PRE PREOG"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PROFESSOR && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-4, 0, 0).Unix()) {
		return pbrc.ReleaseMetadata_PRE_DISTINGUISHED, NO_CHANGE, "PRE DISTIN"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_ASSESS {
		return pbrc.ReleaseMetadata_PRE_FRESHMAN, NO_CHANGE, "ASSESSED"
	}

	return pbrc.ReleaseMetadata_UNKNOWN, NO_CHANGE, "No rules applied"
}
