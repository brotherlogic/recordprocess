package main

import (
	"errors"
	"testing"
	"time"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
)

type testGetter struct {
	lastCategory *pbrc.ReleaseMetadata_Category
	rec          *pbrc.Record
}

func (t *testGetter) getRecords() ([]*pbrc.Record, error) {
	return []*pbrc.Record{t.rec}, nil
}

func (t *testGetter) update(r *pbrc.Record) error {
	t.lastCategory = &r.GetMetadata().Category
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

	return s
}

func TestUpdate(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 1}}}
	s.getter = &tg
	s.processRecords()

	if *tg.lastCategory != pbrc.ReleaseMetadata_PURCHASED {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToUnlistened(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().Unix()}}}
	s.getter = &tg
	s.processRecords()

	if *tg.lastCategory != pbrc.ReleaseMetadata_UNLISTENED {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToStaged(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().Unix()}}}
	s.getter = &tg
	s.processRecords()

	if *tg.lastCategory != pbrc.ReleaseMetadata_STAGED {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToPreFreshman(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812, Rating: 0}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().AddDate(0, -4, 0).Unix()}}}
	s.getter = &tg
	s.processRecords()

	if *tg.lastCategory != pbrc.ReleaseMetadata_PRE_FRESHMAN {
		t.Errorf("Folder has not been updated: %v", tg.lastCategory)
	}
}

func TestUpdateToFreshman(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812, Rating: 4}, Metadata: &pbrc.ReleaseMetadata{DateAdded: time.Now().AddDate(0, -4, 0).Unix()}}}
	s.getter = &tg
	s.processRecords()

	if *tg.lastCategory != pbrc.ReleaseMetadata_FRESHMAN {
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
