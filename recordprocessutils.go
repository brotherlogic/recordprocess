package main

import (
	"fmt"
	"time"

	pbrc "github.com/brotherlogic/recordcollection/proto"
)

type getter interface {
	getRecords() ([]*pbrc.Record, error)
	update(*pbrc.Record) error
}

func (s *Server) processRecords() {
	t := time.Now()
	records, err := s.getter.getRecords()

	if err != nil {
		s.Log(fmt.Sprintf("Error processing records: %v", err))
		return
	}

	s.Log(fmt.Sprintf("About to process %v records", len(records)))
	count := int64(0)
	for _, record := range records {
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
	s.Log(fmt.Sprintf("Processed %v records (touched %v) in %v", len(records), count, time.Now().Sub(t)))
}

func (s *Server) processRecord(r *pbrc.Record) *pbrc.Record {
	if r.GetMetadata() == nil {
		r.Metadata = &pbrc.ReleaseMetadata{}
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

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_FRESHMAN {
		if r.GetMetadata().GetDateAdded() > (time.Now().AddDate(0, -6, 0).Unix()) && r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -3, 0).Unix()) {
			if r.GetRelease().Rating == 0 {
				r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_FRESHMAN
				return r
			}
			r.GetMetadata().Category = pbrc.ReleaseMetadata_FRESHMAN
			return r
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_PROFESSOR {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-3, 0, 0).Unix()) {
			if r.GetRelease().Rating == 0 {
				r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_PROFESSOR
				return r
			}
			r.GetMetadata().Category = pbrc.ReleaseMetadata_PROFESSOR
			return r
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_POSTDOC {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-2, 0, 0).Unix()) {
			if r.GetRelease().Rating == 0 {
				r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_POSTDOC
				return r
			}
			r.GetMetadata().Category = pbrc.ReleaseMetadata_POSTDOC
			return r
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_GRADUATE {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(-1, 0, 0).Unix()) {
			if r.GetRelease().Rating == 0 {
				r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_GRADUATE
				return r
			}
			r.GetMetadata().Category = pbrc.ReleaseMetadata_GRADUATE
			return r
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_SOPHMORE {
		if r.GetMetadata().GetDateAdded() < (time.Now().AddDate(0, -6, 0).Unix()) {
			if r.GetRelease().Rating == 0 {
				r.GetMetadata().Category = pbrc.ReleaseMetadata_PRE_SOPHMORE
				return r
			}
			r.GetMetadata().Category = pbrc.ReleaseMetadata_SOPHMORE
			return r
		}
	}

	return nil
}
