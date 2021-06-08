package main

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/brotherlogic/goserver/utils"
	keystoreclient "github.com/brotherlogic/keystore/client"
	pb "github.com/brotherlogic/recordprocess/proto"
	"golang.org/x/net/context"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
)

type testGetter struct {
	lastCategory pbrc.ReleaseMetadata_Category
	rec          *pbrc.Record
	sold         *pbrc.Record
	getFail      bool
	repeat       int
}

func (t *testGetter) getRecords(ctx context.Context, ti int64, c int) ([]int32, error) {
	if t.repeat > 0 {
		ret := []int32{}
		for i := 0; i < t.repeat; i++ {
			ret = append(ret, t.rec.GetRelease().InstanceId)
		}
		return ret, nil
	}
	return []int32{t.rec.GetRelease().InstanceId}, nil
}

func (t *testGetter) getRecord(ctx context.Context, instanceID int32) (*pbrc.Record, error) {
	if t.getFail {
		return nil, fmt.Errorf("Built to fail")
	}
	return t.rec, nil
}

func (t *testGetter) update(ctx context.Context, instanceID int32, cat pbrc.ReleaseMetadata_Category, reason string, scount int32) error {
	t.lastCategory = cat
	return nil
}

func (t *testGetter) updateStock(ctx context.Context, rec *pbrc.Record) error {
	return nil
}

func (t *testGetter) moveToSold(ctx context.Context, r *pbrc.Record) {
	t.sold = r
}

type testFailGetter struct {
	grf          bool
	lastCategory pbrc.ReleaseMetadata_Category
}

func (t testFailGetter) getRecords(ctx context.Context, ti int64, c int) ([]int32, error) {
	if t.grf {
		return []int32{1}, nil
	}
	return nil, errors.New("Built to fail")
}

func (t testFailGetter) getRecord(ctx context.Context, instanceID int32) (*pbrc.Record, error) {
	if t.grf {
		return &pbrc.Record{}, nil
	}
	return nil, errors.New("Built to fail")
}

func (t testFailGetter) update(ctx context.Context, instanceID int32, cat pbrc.ReleaseMetadata_Category, reason string, blah int32) error {
	if !t.grf {
		t.lastCategory = cat
		return nil
	}
	return errors.New("Built to fail")
}

func (t testFailGetter) updateStock(ctx context.Context, rec *pbrc.Record) error {
	return nil
}

func (t testFailGetter) moveToSold(ctx context.Context, r *pbrc.Record) {
	// Pass
}

func InitTest() *Server {
	s := Init()
	s.SkipLog = true
	s.SkipIssue = true
	s.getter = &testGetter{}
	s.GoServer.KSclient = *keystoreclient.GetTestClient(".testing")
	s.GoServer.KSclient.Save(context.Background(), KEY, &pb.Scores{})

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
		_, _, appl := s.processRecord(context.Background(), rec)

		if utils.FuzzyMatch(rec, test.out) != nil {
			t.Errorf("MATCH FAIL: %v", utils.FuzzyMatch(rec, test.out))
			t.Errorf("Full Test move failed \n%v\n %v \n%v\n (should have been %v)", utils.FuzzyMatch(rec, test.out), rec, appl, test.out)
		}
	}
}

var movetests = []struct {
	in  *pbrc.Record
	out pbrc.ReleaseMetadata_Category
}{
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1362206, Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_ASSESS, Purgatory: pbrc.Purgatory_NEEDS_STOCK_CHECK, LastStockCheck: time.Now().Unix(), GoalFolder: 242017, Cost: 12}}, pbrc.ReleaseMetadata_PRE_FRESHMAN},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1727264, Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PARENTS, GoalFolder: 242017, Cost: 12}}, pbrc.ReleaseMetadata_PRE_FRESHMAN},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_LISTED_TO_SELL, Cost: 12, GoalFolder: 242017}}, pbrc.ReleaseMetadata_SALE_ISSUE},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_LISTED_TO_SELL, SaleId: -1, Cost: 12, GoalFolder: 242017}}, pbrc.ReleaseMetadata_UNLISTENED},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PURCHASED, Cost: 12, GoalFolder: 268147}}, pbrc.ReleaseMetadata_DIGITAL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_BANDCAMP, Cost: 12, GoalFolder: 1782105}}, pbrc.ReleaseMetadata_UNLISTENED},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PURCHASED, Cost: 12}}, pbrc.ReleaseMetadata_UNLISTENED},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1112, Rating: 0, Formats: []*pbgd.Format{&pbgd.Format{Name: "12"}}, Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PROFESSOR, SaleId: 123}}, pbrc.ReleaseMetadata_LISTED_TO_SELL},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1112, Rating: 0, Formats: []*pbgd.Format{&pbgd.Format{Name: "CD"}}, Labels: []*pbgd.Label{&pbgd.Label{Name: "blah"}}}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PREPARE_TO_SELL, LastStockCheck: time.Now().Unix()}}, pbrc.ReleaseMetadata_RIP_THEN_SELL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1111, Rating: 0}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_DIGITAL}}, pbrc.ReleaseMetadata_UNKNOWN},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1235, Rating: 0}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_DIGITAL, GoalFolder: 1235}}, pbrc.ReleaseMetadata_ASSESS},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1234, Rating: 0}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PREPARE_TO_SELL, LastStockCheck: time.Now().Unix()}}, pbrc.ReleaseMetadata_STAGED_TO_SELL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1433217, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNKNOWN}}, pbrc.ReleaseMetadata_GOOGLE_PLAY},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1727264, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNKNOWN}}, pbrc.ReleaseMetadata_PARENTS},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1236, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_STAGED_TO_SELL}}, pbrc.ReleaseMetadata_IN_COLLECTION},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1237, Rating: 3}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_STAGED_TO_SELL}}, pbrc.ReleaseMetadata_SOLD},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1238, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{SetRating: 5}}, pbrc.ReleaseMetadata_UNKNOWN},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1240}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_STAGED, DateAdded: time.Now().AddDate(0, -3, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 12402, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL, DateAdded: time.Now().AddDate(0, -3, 0).Unix()}}, pbrc.ReleaseMetadata_HIGH_SCHOOL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 12403}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_HIGH_SCHOOL, DateAdded: time.Now().AddDate(0, -4, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_IN_COLLECTION},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 12412, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_FRESHMAN, DateAdded: time.Now().AddDate(0, -4, 0).Unix()}}, pbrc.ReleaseMetadata_IN_COLLECTION},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 12413, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_FRESHMAN, DateAdded: time.Now().AddDate(-1, 0, 0).Unix()}}, pbrc.ReleaseMetadata_IN_COLLECTION},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1241}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_FRESHMAN, DateAdded: time.Now().AddDate(0, -7, 0).Unix()}}, pbrc.ReleaseMetadata_IN_COLLECTION},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1242, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_SOPHMORE, DateAdded: time.Now().AddDate(0, -7, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_IN_COLLECTION},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 12435, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_SOPHMORE, DateAdded: time.Now().AddDate(0, -13, 0).Unix()}}, pbrc.ReleaseMetadata_IN_COLLECTION},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1244, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_GRADUATE, DateAdded: time.Now().AddDate(0, -13, 0).Unix()}}, pbrc.ReleaseMetadata_GRADUATE},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1245, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_GRADUATE, DateAdded: time.Now().AddDate(-2, -1, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_POSTDOC},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1242, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_POSTDOC, DateAdded: time.Now().AddDate(-2, -1, 0).Unix()}}, pbrc.ReleaseMetadata_POSTDOC},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 12436, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_POSTDOC, DateAdded: time.Now().AddDate(-3, -1, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_PROFESSOR},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1242, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_PROFESSOR, DateAdded: time.Now().AddDate(-3, -1, 0).Unix()}}, pbrc.ReleaseMetadata_PROFESSOR},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1243, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PROFESSOR, DateAdded: time.Now().AddDate(-4, -1, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_DISTINGUISHED},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1239, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_DISTINGUISHED, DateAdded: time.Now().AddDate(-10, 0, 0).Unix()}}, pbrc.ReleaseMetadata_DISTINGUISHED},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1246, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{FilePath: "1234", Category: pbrc.ReleaseMetadata_RIP_THEN_SELL, DateAdded: time.Now().AddDate(-2, -1, 0).Unix()}}, pbrc.ReleaseMetadata_PREPARE_TO_SELL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1247, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{SaleId: 1234, FilePath: "1234", Category: pbrc.ReleaseMetadata_SOLD, DateAdded: time.Now().AddDate(-2, -1, 0).Unix()}}, pbrc.ReleaseMetadata_LISTED_TO_SELL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1249, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{SaleId: 1234, FilePath: "1234", Category: pbrc.ReleaseMetadata_ASSESS_FOR_SALE, DateAdded: time.Now().AddDate(-2, -1, 0).Unix(), SalePrice: 1234, LastStockCheck: time.Now().Unix()}}, pbrc.ReleaseMetadata_PREPARE_TO_SELL},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1249, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{SaleId: 1234, FilePath: "1234", Category: pbrc.ReleaseMetadata_LISTED_TO_SELL, SaleState: pbgd.SaleState_SOLD, DateAdded: time.Now().AddDate(-2, -1, 0).Unix(), SalePrice: 1234, LastStockCheck: time.Now().Unix()}}, pbrc.ReleaseMetadata_SOLD_ARCHIVE},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1244, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_FRESHMAN, DateAdded: time.Now().AddDate(0, -1, 0).Unix()}}, pbrc.ReleaseMetadata_UNLISTENED},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{SaleId: -1, Category: pbrc.ReleaseMetadata_POSTDOC, DateAdded: time.Now().AddDate(-2, -1, 0).Unix()}}, pbrc.ReleaseMetadata_PURCHASED},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 12, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNLISTENED, DateAdded: time.Now().AddDate(-2, -1, 0).Unix()}}, pbrc.ReleaseMetadata_STAGED},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1242, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_VALIDATE, DateAdded: time.Now().AddDate(0, -7, 0).Unix()}}, pbrc.ReleaseMetadata_VALIDATE},
	{&pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 124334, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_IN_COLLECTION, DateAdded: time.Now().AddDate(0, -13, 0).Unix()}}, pbrc.ReleaseMetadata_IN_COLLECTION},
}

func TestMoveTests(t *testing.T) {
	for _, test := range movetests {
		test.in.GetRelease().InstanceId = 123
		s := InitTest()
		tg := testGetter{rec: test.in}
		s.getter = &tg

		s.ClientUpdate(context.Background(), &pbrc.ClientUpdateRequest{InstanceId: 123})

		if tg.lastCategory != test.out {
			t.Errorf("Full move failed \n%v\n vs. \n%v\n (should have been %v (from %v))", test.in, tg.lastCategory, test.out, tg.rec)
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
	if !val2 {
		t.Errorf("Second save did not fail")
	}
}

func TestUpdate(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{}, Release: &pbgd.Release{InstanceId: 123, Id: 10, Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1}}}
	s.getter = &tg

	s.ClientUpdate(context.Background(), &pbrc.ClientUpdateRequest{InstanceId: 123})

	if tg.lastCategory != pbrc.ReleaseMetadata_PURCHASED {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestBigUpdate(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{}, Release: &pbgd.Release{InstanceId: 123, Id: 10, Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1}}, repeat: 200}
	s.getter = &tg
	s.ClientUpdate(context.Background(), &pbrc.ClientUpdateRequest{InstanceId: 123})
	//s.processNextRecords(context.Background())

	//s.processRecords(context.Background())

	if tg.lastCategory != pbrc.ReleaseMetadata_PURCHASED {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateWithFailonRecordGet(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{}, Release: &pbgd.Release{Id: 10, Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1}}}
	tg.getFail = true
	s.getter = &tg

	//err := s.processNextRecords(context.Background())

	//if err == nil {
	//	t.Errorf("Did not fail")
	//}
}

func TestMultiUpdate(t *testing.T) {
	s := InitTest()

	for i := 0; i < 100; i++ {
		tg := testGetter{rec: &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{}, Release: &pbgd.Release{InstanceId: 10, Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1}}}
		s.getter = &tg
		//s.processNextRecords(context.Background())

		//s.processRecords(context.Background())
	}

	tg := testGetter{rec: &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{}, Release: &pbgd.Release{InstanceId: 11, Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1}}}
	s.getter = &tg
	//s.processNextRecords(context.Background())

	//s.processRecords(context.Background())

	//if s.updateCount != 0 {
	//	t.Errorf("Error in update count: %v", s.updateCount)
	//}

}

func TestUpdateFailOnGet(t *testing.T) {
	s := InitTest()
	tg := testFailGetter{}
	s.getter = tg
	//s.processRecords(context.Background())

	if tg.lastCategory == pbrc.ReleaseMetadata_PURCHASED {
		t.Errorf("Folder has been updated: %v", tg.lastCategory)
	}
}

func TestProcessUnpurchasedRecord(t *testing.T) {
	s := InitTest()
	r := &pbrc.Record{Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1}}
	nr, _, _ := s.processRecord(context.Background(), r)

	if nr != pbrc.ReleaseMetadata_PURCHASED {
		t.Fatalf("Error in processing record: %v", nr)
	}
}

func TestEmptyUpdate(t *testing.T) {
	s := InitTest()
	r := &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PURCHASED}, Release: &pbgd.Release{Labels: []*pbgd.Label{&pbgd.Label{Name: "Label"}}, FolderId: 1}}
	nr, _, _ := s.processRecord(context.Background(), r)

	if nr != pbrc.ReleaseMetadata_UNKNOWN {
		t.Fatalf("Error in processing record: %v", nr)
	}
}

func TestEmptyUpdateOnly(t *testing.T) {
	//s := InitTest()
	//nr := s.processNextRecords(context.Background())

	//if nr != nil {
	//	t.Fatalf("Error in processing record: %v", nr)
	//}
}

func TestIsJustCd(t *testing.T) {
	s := InitTest()
	rec := &pbrc.Record{Release: &pbgd.Release{Formats: []*pbgd.Format{&pbgd.Format{Name: "CD"}}}}
	if !s.isJustCd(context.Background(), rec) {
		t.Errorf("Bad record")
	}
}

func TestIsNotJustCd(t *testing.T) {
	s := InitTest()
	rec := &pbrc.Record{Release: &pbgd.Release{Formats: []*pbgd.Format{&pbgd.Format{Name: "CD"}, &pbgd.Format{Name: "LP"}}}}
	if s.isJustCd(context.Background(), rec) {
		t.Errorf("Bad record")
	}
}
