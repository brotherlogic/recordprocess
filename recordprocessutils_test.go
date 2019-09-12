package main

import (
	"errors"
	"log"
	"testing"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordprocess/proto"
)

type testGetter struct {
	lastCategory pbrc.ReleaseMetadata_Category
	rec          *pbrc.Record
	sold         *pbrc.Record
}

func (t *testGetter) getRecords(ctx context.Context) ([]*pbrc.Record, error) {
	return []*pbrc.Record{t.rec}, nil
}

func (t *testGetter) update(ctx context.Context, r *pbrc.Record) error {
	t.lastCategory = r.GetMetadata().Category
	return nil
}

func (t *testGetter) moveToSold(ctx context.Context, r *pbrc.Record) {
	t.sold = r
}

type testFailGetter struct {
	grf          bool
	lastCategory pbrc.ReleaseMetadata_Category
}

func (t testFailGetter) getRecords(ctx context.Context) ([]*pbrc.Record, error) {
	if t.grf {
		return []*pbrc.Record{&pbrc.Record{Release: &pbgd.Release{FolderId: 1}}}, nil
	}
	return nil, errors.New("Built to fail")
}

func (t testFailGetter) update(ctx context.Context, r *pbrc.Record) error {
	if !t.grf {
		t.lastCategory = r.GetMetadata().GetCategory()
		return nil
	}
	return errors.New("Built to fail")
}

func (t testFailGetter) moveToSold(ctx context.Context, r *pbrc.Record) {
	// Pass
}

func InitTest() *Server {
	s := Init()
	s.SkipLog = true
	s.getter = &testGetter{}
	s.scores = &pb.Scores{}
	s.GoServer.KSclient = *keystoreclient.GetTestClient(".testing")

	return s
}

var fulltests = []struct {
	in  *pbrc.Record
	out *pbrc.Record
}{
	{&pbrc.Record{Release: &pbgd.Release{Formats: []*pbgd.Format{&pbgd.Format{Name: "CD"}}, Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}},
		Metadata: &pbrc.ReleaseMetadata{DateAdded: 1369672260, Others: true, LastCache: 1552578483, Category: pbrc.ReleaseMetadata_RIP_THEN_SELL, GoalFolder: 267115, LastSyncTime: 1561989524, LastStockCheck: 1561971732, OverallScore: 4.75, InstanceId: 19868005, Keep: pbrc.ReleaseMetadata_NOT_KEEPER, Match: pbrc.ReleaseMetadata_FULL_MATCH, CurrentSalePrice: 2241, SalePriceUpdate: 1559853487}},
		&pbrc.Record{Release: &pbgd.Release{Formats: []*pbgd.Format{&pbgd.Format{Name: "CD"}}, Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}},
			Metadata: &pbrc.ReleaseMetadata{DateAdded: 1369672260, Others: true, LastCache: 1552578483, Category: pbrc.ReleaseMetadata_RIP_THEN_SELL, GoalFolder: 267115, LastSyncTime: 1561989524, LastStockCheck: 1561971732, OverallScore: 4.75, InstanceId: 19868005, Keep: pbrc.ReleaseMetadata_NOT_KEEPER, Match: pbrc.ReleaseMetadata_FULL_MATCH, CurrentSalePrice: 2241, SalePriceUpdate: 1559853487}},
	},
}

func TestFullTests(t *testing.T) {
	for _, test := range fulltests {
		s := InitTest()
		rec := test.in
		tg := testGetter{rec: test.in}
		s.getter = &tg
		_, appl := s.processRecord(rec)

		if !utils.FuzzyMatch(rec, test.out) {
			t.Errorf("Full Test move failed \n%v\n %v \n%v\n (should have been %v)", rec, appl, test.out)
		}
	}
}

var movetests = []struct {
	in  *pbrc.Record
	out pbrc.ReleaseMetadata_Category
}{
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_LISTED_TO_SELL, SaleId: -1, Cost: 12, GoalFolder: 242017}}, pbrc.ReleaseMetadata_UNLISTENED},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PURCHASED, Cost: 12, GoalFolder: 268147}}, pbrc.ReleaseMetadata_DIGITAL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PURCHASED, Cost: 12}}, pbrc.ReleaseMetadata_UNLISTENED},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1112, Rating: 0, Formats: []*pbgd.Format{&pbgd.Format{Name: "12"}}, Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PROFESSOR, SaleId: 123}}, pbrc.ReleaseMetadata_LISTED_TO_SELL},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1112, Rating: 0, Formats: []*pbgd.Format{&pbgd.Format{Name: "CD"}}, Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PREPARE_TO_SELL, LastStockCheck: time.Now().Unix()}}, pbrc.ReleaseMetadata_RIP_THEN_SELL},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1111, Rating: 0}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_GRADUATE}}, pbrc.ReleaseMetadata_NO_LABELS},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1111, Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}, Rating: 0}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_NO_LABELS}}, pbrc.ReleaseMetadata_PRE_FRESHMAN},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1111, Rating: 0}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_DIGITAL}}, pbrc.ReleaseMetadata_UNKNOWN},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1235, Rating: 0}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_DIGITAL, GoalFolder: 1235}}, pbrc.ReleaseMetadata_ASSESS},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1234, Rating: 0}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PREPARE_TO_SELL, LastStockCheck: time.Now().Unix()}}, pbrc.ReleaseMetadata_STAGED_TO_SELL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1433217, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNKNOWN}}, pbrc.ReleaseMetadata_GOOGLE_PLAY},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1727264, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNKNOWN}}, pbrc.ReleaseMetadata_PARENTS},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1236, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_STAGED_TO_SELL}}, pbrc.ReleaseMetadata_PRE_FRESHMAN},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1237, Rating: 3}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_STAGED_TO_SELL}}, pbrc.ReleaseMetadata_SOLD},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1238, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{SetRating: 5}}, pbrc.ReleaseMetadata_UNKNOWN},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1239, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_FRESHMAN, DateAdded: time.Now().AddDate(-10, 0, 0).Unix()}}, pbrc.ReleaseMetadata_DISTINGUISHED},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1240}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_STAGED, DateAdded: time.Now().AddDate(0, -3, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1240, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL, DateAdded: time.Now().AddDate(0, -3, 0).Unix()}}, pbrc.ReleaseMetadata_HIGH_SCHOOL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1240}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_HIGH_SCHOOL, DateAdded: time.Now().AddDate(0, -4, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_FRESHMAN},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1241}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_FRESHMAN, DateAdded: time.Now().AddDate(0, -7, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_SOPHMORE},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1242, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_SOPHMORE, DateAdded: time.Now().AddDate(0, -7, 0).Unix()}}, pbrc.ReleaseMetadata_SOPHMORE},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1243, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_SOPHMORE, DateAdded: time.Now().AddDate(0, -13, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_GRADUATE},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1244, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_GRADUATE, DateAdded: time.Now().AddDate(0, -13, 0).Unix()}}, pbrc.ReleaseMetadata_GRADUATE},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1245, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_GRADUATE, DateAdded: time.Now().AddDate(-2, -1, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_POSTDOC},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1243, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_POSTDOC, DateAdded: time.Now().AddDate(-3, -1, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_PROFESSOR},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1243, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PROFESSOR, DateAdded: time.Now().AddDate(-4, -1, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_DISTINGUISHED},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1246, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{FilePath: "1234", Category: pbrc.ReleaseMetadata_RIP_THEN_SELL, DateAdded: time.Now().AddDate(-2, -1, 0).Unix()}}, pbrc.ReleaseMetadata_PREPARE_TO_SELL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1247, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{SaleId: 1234, FilePath: "1234", Category: pbrc.ReleaseMetadata_SOLD, DateAdded: time.Now().AddDate(-2, -1, 0).Unix()}}, pbrc.ReleaseMetadata_LISTED_TO_SELL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1249, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{SaleId: 1234, FilePath: "1234", Category: pbrc.ReleaseMetadata_SOLD, DateAdded: time.Now().AddDate(-2, -1, 0).Unix(), SalePrice: 1234}}, pbrc.ReleaseMetadata_SOLD_ARCHIVE},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1249, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{SaleId: 1234, FilePath: "1234", Category: pbrc.ReleaseMetadata_PREPARE_TO_SELL, DateAdded: time.Now().AddDate(-2, -1, 0).Unix(), SalePrice: 1234}}, pbrc.ReleaseMetadata_ASSESS_FOR_SALE},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1249, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{SaleId: 1234, FilePath: "1234", Category: pbrc.ReleaseMetadata_ASSESS_FOR_SALE, DateAdded: time.Now().AddDate(-2, -1, 0).Unix(), SalePrice: 1234, LastStockCheck: time.Now().Unix()}}, pbrc.ReleaseMetadata_PREPARE_TO_SELL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1249, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{SaleId: 1234, FilePath: "1234", Category: pbrc.ReleaseMetadata_LISTED_TO_SELL, SaleState: pbgd.SaleState_SOLD, DateAdded: time.Now().AddDate(-2, -1, 0).Unix(), SalePrice: 1234, LastStockCheck: time.Now().Unix()}}, pbrc.ReleaseMetadata_SOLD_ARCHIVE},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1244, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_FRESHMAN, DateAdded: time.Now().AddDate(0, -1, 0).Unix()}}, pbrc.ReleaseMetadata_UNLISTENED},
}

func TestMoveTests(t *testing.T) {
	for _, test := range movetests {
		log.Printf("TESTING: %v", test.in)
		s := InitTest()
		tg := testGetter{rec: test.in}
		s.getter = &tg
		s.processRecords(context.Background())

		if tg.lastCategory != test.out {
			t.Fatalf("Full move failed \n%v\n vs. \n%v\n (should have been %v (from %v))", test.in, tg.lastCategory, test.out, tg.rec)
		}
	}
}

func TestSaveRecordTwice(t *testing.T) {
	s := InitTest()
	val := s.saveRecordScore(context.Background(), &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, InstanceId: 1234, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_FRESHMAN}})

	if !val {
		t.Fatalf("First save failed")
	}

	val2 := s.saveRecordScore(context.Background(), &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, InstanceId: 1234, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_FRESHMAN}})
	if val2 {
		t.Errorf("Second save did not fail")
	}
}

func TestUpdate(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{}, Release: &pbgd.Release{Id: 10, Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1}}}
	s.getter = &tg
	s.processRecords(context.Background())

	if tg.lastCategory != pbrc.ReleaseMetadata_PURCHASED {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestMultiUpdate(t *testing.T) {
	s := InitTest()

	for i := 0; i < 100; i++ {
		tg := testGetter{rec: &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{}, Release: &pbgd.Release{InstanceId: 10, Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1}}}
		s.getter = &tg
		s.processRecords(context.Background())
	}

	tg := testGetter{rec: &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{}, Release: &pbgd.Release{InstanceId: 11, Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1}}}
	s.getter = &tg
	s.processRecords(context.Background())

	if s.updateCount != 0 {
		t.Errorf("Error in update count: %v", s.updateCount)
	}

}

func TestUpdateToUnlistened(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().Unix()}}}
	s.getter = &tg
	s.processRecords(context.Background())

	if tg.lastCategory != pbrc.ReleaseMetadata_UNLISTENED {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToStaged(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().Unix()}}}
	s.getter = &tg
	s.processRecords(context.Background())

	if tg.lastCategory != pbrc.ReleaseMetadata_STAGED {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToFreshman(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().AddDate(0, -4, 0).Unix()}}}
	s.getter = &tg
	s.processRecords(context.Background())

	if tg.lastCategory != pbrc.ReleaseMetadata_FRESHMAN {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToProfessor(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().AddDate(-3, -1, 0).Unix()}}}
	s.getter = &tg
	s.processRecords(context.Background())

	if tg.lastCategory != pbrc.ReleaseMetadata_PROFESSOR {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToPostdoc(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().AddDate(-2, -1, 0).Unix()}}}
	s.getter = &tg
	s.processRecords(context.Background())

	if tg.lastCategory != pbrc.ReleaseMetadata_POSTDOC {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToGraduate(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().AddDate(-1, -1, 0).Unix()}}}
	s.getter = &tg
	s.processRecords(context.Background())

	if tg.lastCategory != pbrc.ReleaseMetadata_GRADUATE {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToSophmore(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().AddDate(0, -7, 0).Unix()}}}
	s.getter = &tg
	s.processRecords(context.Background())

	if tg.lastCategory != pbrc.ReleaseMetadata_SOPHMORE {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateFailOnGet(t *testing.T) {
	s := InitTest()
	tg := testFailGetter{}
	s.getter = tg
	s.processRecords(context.Background())

	if tg.lastCategory == pbrc.ReleaseMetadata_PURCHASED {
		t.Errorf("Folder has been updated: %v", tg.lastCategory)
	}
}

func TestUpdateFailOnUpdate(t *testing.T) {
	s := InitTest()
	tg := testFailGetter{grf: true}
	s.getter = tg
	s.processRecords(context.Background())

	if tg.lastCategory == pbrc.ReleaseMetadata_PURCHASED {
		t.Errorf("Folder has been updated: %v", tg.lastCategory)
	}
}

func TestProcessUnpurchasedRecord(t *testing.T) {
	s := InitTest()
	r := &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1}}
	nr, _ := s.processRecord(r)

	if nr.GetMetadata().GetCategory() != pbrc.ReleaseMetadata_PURCHASED {
		t.Fatalf("Error in processing record: %v", nr)
	}
}

func TestEmptyUpdate(t *testing.T) {
	s := InitTest()
	r := &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PURCHASED}, Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1}}
	nr, _ := s.processRecord(r)

	if nr != nil {
		t.Fatalf("Error in processing record: %v", nr)
	}
}

func TestPromoteToStaged(t *testing.T) {
	s := InitTest()
	rec := &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 812802, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNLISTENED, DateAdded: time.Now().Unix()}}
	tg := testGetter{rec: rec}
	s.getter = &tg
	s.processRecords(context.Background())

	if rec.GetMetadata().GetCategory() != pbrc.ReleaseMetadata_STAGED {
		t.Errorf("Folder has not been updated: %v", rec)
	}
}

func TestClearRatingOnPrepToSell(t *testing.T) {
	s := InitTest()
	rec := &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 812802, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PREPARE_TO_SELL, DateAdded: time.Now().Unix(), LastStockCheck: time.Now().Unix()}}
	tg := testGetter{rec: rec}
	s.getter = &tg
	s.processRecords(context.Background())

	if rec.GetMetadata().SetRating != -1 {
		t.Errorf("Folder has not been updated: %v", rec)
	}
}

func TestClearRatingOnStagedToSell(t *testing.T) {
	s := InitTest()
	rec := &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 812802, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_STAGED_TO_SELL, DateAdded: time.Now().Unix(), LastStockCheck: time.Now().Unix()}}
	tg := testGetter{rec: rec}
	s.getter = &tg
	s.processRecords(context.Background())

	if rec.GetMetadata().SetRating != -1 {
		t.Errorf("Folder has not been updated: %v", rec)
	}
}

func TestBandcamp(t *testing.T) {
	s := InitTest()
	rec := &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1782105, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PREPARE_TO_SELL, DateAdded: time.Now().Unix(), LastStockCheck: time.Now().Unix()}}
	tg := testGetter{rec: rec}
	s.getter = &tg
	s.processRecords(context.Background())

	if rec.GetMetadata().GoalFolder != 1782105 {
		t.Errorf("Folder has not been updated: %v", rec)
	}
}
