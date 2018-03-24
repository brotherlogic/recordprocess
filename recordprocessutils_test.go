package main

import (
	"errors"
	"testing"
	"time"

	"github.com/brotherlogic/keystore/client"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordprocess/proto"
)

type testGetter struct {
	lastCategory pbrc.ReleaseMetadata_Category
	rec          *pbrc.Record
}

func (t *testGetter) getRecords() ([]*pbrc.Record, error) {
	return []*pbrc.Record{t.rec}, nil
}

func (t *testGetter) update(r *pbrc.Record) error {
	t.lastCategory = r.GetMetadata().Category
	return nil
}

type testFailGetter struct {
	grf          bool
	lastCategory pbrc.ReleaseMetadata_Category
}

func (t testFailGetter) getRecords() ([]*pbrc.Record, error) {
	if t.grf {
		return []*pbrc.Record{&pbrc.Record{Release: &pbgd.Release{FolderId: 1}}}, nil
	}
	return nil, errors.New("Built to fail")
}

func (t testFailGetter) update(r *pbrc.Record) error {
	if !t.grf {
		t.lastCategory = r.GetMetadata().GetCategory()
		return nil
	}
	return errors.New("Built to fail")
}

func InitTest() *Server {
	s := Init()
	s.SkipLog = true
	s.getter = &testGetter{}
	s.scores = &pb.Scores{}
	s.GoServer.KSclient = *keystoreclient.GetTestClient(".testing")

	return s
}

var movetests = []struct {
	in  *pbrc.Record
	out pbrc.ReleaseMetadata_Category
}{
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1111, Rating: 0}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_DIGITAL}}, pbrc.ReleaseMetadata_UNKNOWN},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1234, Rating: 0}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_DIGITAL, GoalFolder: 1235}}, pbrc.ReleaseMetadata_ASSESS},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1234, Rating: 0}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PREPARE_TO_SELL}}, pbrc.ReleaseMetadata_STAGED_TO_SELL},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1433217, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNKNOWN}}, pbrc.ReleaseMetadata_GOOGLE_PLAY},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1234, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_STAGED_TO_SELL}}, pbrc.ReleaseMetadata_PRE_FRESHMAN},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1234, Rating: 3}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_STAGED_TO_SELL}}, pbrc.ReleaseMetadata_SOLD},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1234, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{SetRating: 5}}, pbrc.ReleaseMetadata_UNKNOWN},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1234, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_FRESHMAN, DateAdded: time.Now().AddDate(-10, 0, 0).Unix()}}, pbrc.ReleaseMetadata_PROFESSOR},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 268147}, Metadata: &pbrc.ReleaseMetadata{Purgatory: pbrc.Purgatory_NEEDS_STOCK_CHECK, LastStockCheck: time.Now().Unix()}}, pbrc.ReleaseMetadata_PRE_FRESHMAN},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1234}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_STAGED, DateAdded: time.Now().AddDate(0, -4, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_FRESHMAN},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1234}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_FRESHMAN, DateAdded: time.Now().AddDate(0, -7, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_SOPHMORE},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1234, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_SOPHMORE, DateAdded: time.Now().AddDate(0, -7, 0).Unix()}}, pbrc.ReleaseMetadata_SOPHMORE},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1234, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_SOPHMORE, DateAdded: time.Now().AddDate(0, -13, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_GRADUATE},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1234, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_GRADUATE, DateAdded: time.Now().AddDate(0, -13, 0).Unix()}}, pbrc.ReleaseMetadata_GRADUATE},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1234, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_GRADUATE, DateAdded: time.Now().AddDate(-2, -1, 0).Unix()}}, pbrc.ReleaseMetadata_PRE_POSTDOC},
}

func TestMoveTests(t *testing.T) {
	for _, test := range movetests {
		s := InitTest()
		tg := testGetter{rec: test.in}
		s.getter = &tg
		s.processRecords()

		if tg.lastCategory != test.out {
			t.Fatalf("Test move failed %v -> %v (should have been %v (from %v))", test.in, tg.lastCategory, test.out, tg.rec)
		}
	}
}

func TestSaveRecordTwice(t *testing.T) {
	s := InitTest()
	val := s.saveRecordScore(&pbrc.Record{Release: &pbgd.Release{InstanceId: 1234, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_FRESHMAN}})

	if !val {
		t.Fatalf("First save failed")
	}

	val2 := s.saveRecordScore(&pbrc.Record{Release: &pbgd.Release{InstanceId: 1234, Rating: 5}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_FRESHMAN}})
	if val2 {
		t.Errorf("Second save did not fail")
	}
}

func TestUpdate(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{}, Release: &pbgd.Release{FolderId: 1}}}
	s.getter = &tg
	s.processRecords()

	if tg.lastCategory != pbrc.ReleaseMetadata_PURCHASED {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToUnlistened(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().Unix()}}}
	s.getter = &tg
	s.processRecords()

	if tg.lastCategory != pbrc.ReleaseMetadata_UNLISTENED {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToStaged(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().Unix()}}}
	s.getter = &tg
	s.processRecords()

	if tg.lastCategory != pbrc.ReleaseMetadata_STAGED {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToFreshman(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().AddDate(0, -4, 0).Unix()}}}
	s.getter = &tg
	s.processRecords()

	if tg.lastCategory != pbrc.ReleaseMetadata_FRESHMAN {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToProfessor(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().AddDate(-3, -1, 0).Unix()}}}
	s.getter = &tg
	s.processRecords()

	if tg.lastCategory != pbrc.ReleaseMetadata_PROFESSOR {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToPostdoc(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().AddDate(-2, -1, 0).Unix()}}}
	s.getter = &tg
	s.processRecords()

	if tg.lastCategory != pbrc.ReleaseMetadata_POSTDOC {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToGraduate(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().AddDate(-1, -1, 0).Unix()}}}
	s.getter = &tg
	s.processRecords()

	if tg.lastCategory != pbrc.ReleaseMetadata_GRADUATE {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToSophmore(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().AddDate(0, -7, 0).Unix()}}}
	s.getter = &tg
	s.processRecords()

	if tg.lastCategory != pbrc.ReleaseMetadata_SOPHMORE {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateFailOnGet(t *testing.T) {
	s := InitTest()
	tg := testFailGetter{}
	s.getter = tg
	s.processRecords()

	if tg.lastCategory == pbrc.ReleaseMetadata_PURCHASED {
		t.Errorf("Folder has been updated: %v", tg.lastCategory)
	}
}

func TestUpdateFailOnUpdate(t *testing.T) {
	s := InitTest()
	tg := testFailGetter{grf: true}
	s.getter = tg
	s.processRecords()

	if tg.lastCategory == pbrc.ReleaseMetadata_PURCHASED {
		t.Errorf("Folder has been updated: %v", tg.lastCategory)
	}
}

func TestProcessUnpurchasedRecord(t *testing.T) {
	s := InitTest()
	r := &pbrc.Record{Release: &pbgd.Release{FolderId: 1}}
	nr := s.processRecord(r)

	if nr.GetMetadata().GetCategory() != pbrc.ReleaseMetadata_PURCHASED {
		t.Fatalf("Error in processing record: %v", nr)
	}
}

func TestEmptyUpdate(t *testing.T) {
	s := InitTest()
	r := &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PURCHASED}, Release: &pbgd.Release{FolderId: 1}}
	nr := s.processRecord(r)

	if nr != nil {
		t.Fatalf("Error in processing record: %v", nr)
	}
}

func TestPromoteToStaged(t *testing.T) {
	s := InitTest()
	rec := &pbrc.Record{Release: &pbgd.Release{FolderId: 812802, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNLISTENED, DateAdded: time.Now().Unix()}}
	tg := testGetter{rec: rec}
	s.getter = &tg
	s.processRecords()

	if rec.GetMetadata().GetCategory() != pbrc.ReleaseMetadata_STAGED {
		t.Errorf("Folder has not been updated: %v", rec)
	}
}

func TestClearRatingOnPrepToSell(t *testing.T) {
	s := InitTest()
	rec := &pbrc.Record{Release: &pbgd.Release{FolderId: 812802, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PREPARE_TO_SELL, DateAdded: time.Now().Unix()}}
	tg := testGetter{rec: rec}
	s.getter = &tg
	s.processRecords()

	if rec.GetMetadata().SetRating != -1 {
		t.Errorf("Folder has not been updated: %v", rec)
	}
}
