package rlog

import (
	"github.com/edsrzf/mmap-go"
	"io"
	"os"
)

var (
	offWidth   uint64 = 4                   // Width of the record's offset (offset is similar id).
	posWidth   uint64 = 8                   // Width of the record's position.
	entryWidth        = offWidth + posWidth // entryWidth can be used to jump straight to the position of an entry given its offset since the position in the file is offset * entWidth.
)

type index struct {
	file      *os.File
	memoryMap mmap.MMap
	size      uint64
}

type pair struct {
	idxOffset   uint32
	storeOffset uint64
}

// newIndex - returns a new instance of the index for the given file. Internally uses a memory-mapped file for improved performance.
// We grow the file to max index size before memory mapping the file, because, once we memory-map a file, we cannot grow.
// Hence, we grow while opening the file, and truncate while closing the file.
// Note: Because of this, this doesn't handle the ungraceful shutdown.
// An index entry contains two fields: the record's offset and its position in the store file.
// offset is stored as uint32s and pos as uint64s.
func newIndex(f *os.File, c LogConfig) (*index, error) {
	idx := &index{file: f}
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	idx.size = uint64(fi.Size())
	if err = os.Truncate(f.Name(), int64(c.MaxIndexBytes)); err != nil {
		return nil, err
	}

	if idx.memoryMap, err = mmap.Map(idx.file, mmap.RDWR, 0); err != nil {
		return nil, err
	}
	return idx, nil
}

// Read returns the associated record's position in the store for the given offset.
//
//  1. When we start our service, the service needs to know the offset to set on the next record appended to the log.
//     The service learns the next record’s offset by looking at the last entry of the index, a simple process of reading the last 12 bytes of the file
//
//  2. The input offset is relative to the segment's base offset. 0 is always the offset of index's first entry, 1 is the second and so on.
//     Relative offsets used to reduce the size of the indexes by storing offsets as uint32s. If we used absolute offsets, we’d have to store the offsets as uint64s and require four more bytes for each entry.
func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	// When -1 passed, return the last offset and position in the index
	if in == -1 {
		out = uint32((i.size / entryWidth) - 1)
	} else {
		out = uint32(in)
	}
	pos = uint64(out) * entryWidth
	if i.size < pos+entryWidth {
		return 0, 0, io.EOF
	}
	out = enc.Uint32(i.memoryMap[pos : pos+offWidth])
	pos = enc.Uint64(i.memoryMap[pos+offWidth : pos+entryWidth])
	return out, pos, nil
}

func (i *index) ReadAll(offset uint64) []pair {
	indices := make([]pair, 0)
	if i.size == 0 {
		return indices
	}
	for pos := offset * entryWidth; pos < i.size; pos += entryWidth {
		idxPair := pair{
			idxOffset:   enc.Uint32(i.memoryMap[pos : pos+offWidth]),
			storeOffset: enc.Uint64(i.memoryMap[pos+offWidth : pos+entryWidth]),
		}
		indices = append(indices, idxPair)
	}
	return indices
}

// Write appends the given offset and position to the index.
func (i *index) Write(off uint32, pos uint64) error {
	if uint64(len(i.memoryMap)) < i.size+entryWidth {
		return io.EOF
	}
	enc.PutUint32(i.memoryMap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.memoryMap[i.size+offWidth:i.size+entryWidth], pos)
	i.size += entryWidth
	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}

func (i *index) Close() error {
	if err := i.memoryMap.Flush(); err != nil {
		return err
	}
	if err := i.file.Sync(); err != nil {
		return err
	}
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	return i.file.Close()
}
