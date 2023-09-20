package rlog

import (
	"fmt"
	"github.com/pradeesh-kumar/go-raft-kv/raft"
	"google.golang.org/protobuf/proto"
	"os"
	"path"
)

type segment struct {
	store                  *store
	index                  *index
	baseOffset, nextOffset uint64
	config                 LogConfig
}

// The segment wraps the index and store types to coordinate operations across the two.
func newSegment(dir string, baseOffset uint64, c LogConfig) (*segment, error) {
	s := &segment{
		baseOffset: baseOffset,
		config:     c,
	}
	var err error
	storeFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".store")),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}
	if s.store, err = newStore(storeFile); err != nil {
		return nil, err
	}
	indexFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".index")),
		os.O_RDWR|os.O_CREATE,
		0644,
	)
	if err != nil {
		return nil, err
	}
	if s.index, err = newIndex(indexFile, c); err != nil {
		return nil, err
	}
	if off, _, err := s.index.Read(-1); err != nil {
		s.nextOffset = baseOffset
	} else {
		s.nextOffset = baseOffset + uint64(off) + 1
	}
	return s, nil
}

func (s *segment) Append(record *raft.Record) (offset uint64, err error) {
	cur := s.nextOffset
	record.Offset = cur
	p, err := proto.Marshal(record)
	if err != nil {
		return 0, err
	}
	_, pos, err := s.store.Append(p)
	if err != nil {
		return 0, err
	}
	if err = s.index.Write(
		// index offsets are relative to base offset
		uint32(s.nextOffset-s.baseOffset),
		pos,
	); err != nil {
		return 0, err
	}
	s.nextOffset++
	return cur, nil
}

func (s *segment) Read(offset uint64) (*raft.Record, error) {
	_, pos, err := s.index.Read(int64(offset - s.baseOffset))
	if err != nil {
		return nil, err
	}
	p, err := s.store.Read(pos)
	if err != nil {
		return nil, err
	}
	record := &raft.Record{}
	err = proto.Unmarshal(p, record)
	return record, err
}

func (s *segment) ReadAllSince(offset uint64) ([]*raft.Record, error) {
	records := make([]*raft.Record, 0)
	indices := s.index.ReadAll(offset - s.baseOffset)
	if len(indices) > 0 {
		for _, idx := range indices {
			recordBytes, err := s.store.Read(idx.storeOffset)
			if err != nil {
				return records, fmt.Errorf("error while reading the record at offset %d: %+v", idx.storeOffset, err)
			}
			record := &raft.Record{}
			err = proto.Unmarshal(recordBytes, record)
			if err != nil {
				return records, fmt.Errorf("error while unmarshalling record at offset %d: %+v", idx.storeOffset, err)
			}
			records = append(records, record)
		}
	}
	return records, nil
}

func (s *segment) ReadBatchSince(offset uint64, batchSize int) ([]*raft.Record, error) {
	records := make([]*raft.Record, 0)
	if batchSize == 0 {
		return records, nil
	}
	indices := s.index.ReadAll(offset - s.baseOffset)
	if len(indices) > 0 {
		for _, idx := range indices {
			recordBytes, err := s.store.Read(idx.storeOffset)
			if err != nil {
				return records, fmt.Errorf("error while reading the record at offset %d: %+v", idx.storeOffset, err)
			}
			record := &raft.Record{}
			err = proto.Unmarshal(recordBytes, record)
			if err != nil {
				return records, fmt.Errorf("error while unmarshalling record at offset %d: %+v", idx.storeOffset, err)
			}
			records = append(records, record)
			batchSize--
			if batchSize == 0 {
				break
			}
		}
	}
	return records, nil
}

// IsMaxed returns whether the segment has reached its max size either by writing too much to the store or the index.
// If you wrote a small number of long logs, then you’d hit the segment bytes limit; if you wrote a lot of small logs, then you’d hit the index bytes limit.
// The log uses this method to know it needs to create a new segment.
func (s *segment) IsMaxed() bool {
	return s.store.size >= s.config.MaxStoreBytes ||
		s.index.size >= s.config.MaxIndexBytes
}

// Close closes both index and store file.
func (s *segment) Close() error {
	if err := s.index.Close(); err != nil {
		return err
	}
	if err := s.store.Close(); err != nil {
		return err
	}
	return nil
}

// Remove closes the segment and removes the index and store files.
func (s *segment) Remove() error {
	if err := s.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.index.Name()); err != nil {
		return err
	}
	if err := os.Remove(s.store.Name()); err != nil {
		return err
	}
	return nil
}

// nearestMultiple returns the nearest and lesser multiple of k in j.
// for example nearestMultiple(9, 4) == 8
func nearestMultiple(j, k uint64) uint64 {
	if j >= 0 {
		return (j / k) * k
	}
	return ((j - k + 1) / k) * k
}
