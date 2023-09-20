package rlog

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8 // length of width of the offset
)

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

// newStore creates a store for the given file.
func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

// Append appends the given bytes to the store. It writes to the buffer instead directly to the file system in order to
// reduce the system calls and improve performance. This first writes the length of the record, so that while reading, we
// will know how much data to read for this record
// Returns
// n - number of bytes written
// pos - position where the store holds the record in its file. This segment will use this position when it creates an
// associated index entry for this record.
func (s *store) Append(record []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pos = s.size
	if err = binary.Write(s.buf, enc, uint64(len(record))); err != nil {
		return 0, 0, err
	}
	bytesWritten, err := s.buf.Write(record)
	if err != nil {
		return 0, 0, err
	}
	bytesWritten += lenWidth
	s.size += uint64(bytesWritten)
	return uint64(bytesWritten), pos, nil
}

// Read reads the record from given position.
// First it flushes the buffer to the memory.
// Finds out how many bytes to read
func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}
	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}
	return b, nil
}

// ReadAt - reads the bytes at given offset
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

// Close flushes the buffer and closes the file.
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return err
	}
	return s.File.Close()
}
