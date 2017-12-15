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

	for _, record := range records {
		update := s.processRecord(record)
		if update != nil {
			err := s.getter.update(update)
			if err != nil {
				s.Log(fmt.Sprintf("Error updating record: %v", err))
			}
		}
	}

	s.Log(fmt.Sprintf("Processed %v records in %v", len(records), time.Now().Sub(t)))
}

func (s *Server) processRecord(r *pbrc.Record) *pbrc.Record {
	if r.GetMetadata() == nil {
		r.Metadata = &pbrc.ReleaseMetadata{}
	}

	if r.GetRelease().FolderId == 1 && r.GetMetadata().Category != pbrc.ReleaseMetadata_PURCHASED {
		r.GetMetadata().Category = pbrc.ReleaseMetadata_PURCHASED
		return r
	}

	return nil
}
