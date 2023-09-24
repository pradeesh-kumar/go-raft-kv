package rlog

import (
	"fmt"
	"github.com/pradeesh-kumar/go-raft-kv/raft"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Log struct {
	mu sync.RWMutex

	Dir    string
	Config LogConfig

	activeSegment *segment
	segments      []*segment
}

func NewLog(dir string, c LogConfig) (*Log, error) {
	if c.MaxStoreBytes == 0 {
		c.MaxStoreBytes = 1024
	}
	if c.MaxIndexBytes == 0 {
		c.MaxIndexBytes = 1024
	}
	l := &Log{
		Dir:    dir,
		Config: c,
	}

	return l, l.setup()
}

func (l *Log) setup() error {
	files, err := os.ReadDir(l.Dir)
	if err != nil {
		return err
	}
	var baseOffsets []uint64
	for _, file := range files {
		offStr := strings.TrimSuffix(
			file.Name(),
			path.Ext(file.Name()),
		)
		off, _ := strconv.ParseUint(offStr, 10, 0)
		baseOffsets = append(baseOffsets, off)
	}
	sort.Slice(baseOffsets, func(i, j int) bool {
		return baseOffsets[i] < baseOffsets[j]
	})
	for i := 0; i < len(baseOffsets); i++ {
		if err = l.newSegment(baseOffsets[i]); err != nil {
			return err
		}
		// baseOffset contains dup for index and store so we skip
		// the dup
		i++
	}
	if l.segments == nil {
		if err = l.newSegment(l.Config.InitialOffset); err != nil {
			return err
		}
	}
	return nil
}

func (l *Log) Append(record *raft.Record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	off, err := l.activeSegment.Append(record)
	if err != nil {
		return 0, err
	}
	if l.activeSegment.IsMaxed() {
		err = l.newSegment(off + 1)
	}
	return off, err
}

func (l *Log) Read(offset uint64) (*raft.Record, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var s *segment
	for _, seg := range l.segments {
		if seg.baseOffset <= offset && offset < seg.nextOffset {
			s = seg
			break
		}
	}
	if s == nil || s.nextOffset <= offset {
		return nil, fmt.Errorf("offset out of range: %d", offset)
	}
	return s.Read(offset)
}

func (l *Log) ReadAllSince(offset uint64) ([]*raft.Record, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var s *segment
	for _, seg := range l.segments {
		if seg.baseOffset <= offset && offset < seg.nextOffset {
			s = seg
			break
		}
	}
	if s == nil || s.nextOffset <= offset {
		return nil, fmt.Errorf("offset out of range: %d", offset)
	}
	return s.ReadAllSince(offset)
}

func (l *Log) ReadBatchSince(offset uint64, batchSize int) (logs []*raft.Record, err error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, seg := range l.segments {
		if seg.baseOffset <= offset && offset < seg.nextOffset {
			logsRead, readErr := seg.ReadBatchSince(offset, batchSize)
			if readErr != nil {
				return nil, readErr
			}
			logs = append(logs, logsRead...)
			if len(logs) >= batchSize {
				return
			}
		}
	}
	return nil, fmt.Errorf("offset out of range: %d", offset)
}

func (l *Log) newSegment(off uint64) error {
	s, err := newSegment(l.Dir, off, l.Config)
	if err != nil {
		return err
	}
	l.segments = append(l.segments, s)
	l.activeSegment = s
	return nil
}

func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, s := range l.segments {
		if err := s.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (l *Log) Remove() error {
	if err := l.Close(); err != nil {
		return err
	}
	return os.RemoveAll(l.Dir)
}

func (l *Log) Reset() error {
	if err := l.Remove(); err != nil {
		return err
	}
	return l.setup()
}

func (l *Log) LowestOffset() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.segments[0].baseOffset
}

func (l *Log) HighestOffset() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	off := l.segments[len(l.segments)-1].nextOffset
	if off == 0 {
		return 0
	}
	return off - 1
}

// Truncate removes all segments whose highest offset is lower than lowest.
// This is to save the disk space. Periodically call Truncate to remove old segments whose data has been processed.
// By then the segment isn't required anymore.
func (l *Log) Truncate(lowest uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	var segments []*segment
	for _, s := range l.segments {
		if s.nextOffset <= lowest+1 {
			if err := s.Remove(); err != nil {
				return err
			}
			continue
		}
		segments = append(segments, s)
	}
	l.segments = segments
	return nil
}

func (l *Log) Reader() io.Reader {
	l.mu.RLock()
	defer l.mu.RUnlock()
	readers := make([]io.Reader, len(l.segments))
	for i, s := range l.segments {
		readers[i] = &originReader{s.store, 0}
	}
	return io.MultiReader(readers...)
}

type originReader struct {
	*store
	off int64
}

func (o *originReader) Read(p []byte) (int, error) {
	n, err := o.ReadAt(p, o.off)
	o.off += int64(n)
	return n, err
}
