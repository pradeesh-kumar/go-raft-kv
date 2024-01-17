package raft

import (
	"errors"
	"fmt"
	"github.com/pradeesh-kumar/go-raft-kv/logger"
	"os"
	"path"
)

type SnapshotFileInfo struct {
	term         uint32
	logIndex     uint64
	snapshotFile *os.File
}

type SnapshotFileReader struct {
	SnapshotFileInfo
}

func newSnapshotFileReader(dir string, term uint32, logIndex uint64) (SnapshotReader, error) {
	fileName := fmt.Sprintf("raft-snapshot_%d_%d.snapshot", term, logIndex)
	file, err := os.Open(path.Join(dir, fileName))
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("the snapshot file %s doesn't exists", fileName)
	}
	return &SnapshotFileReader{SnapshotFileInfo{term: term, logIndex: logIndex, snapshotFile: file}}, nil
}

func (s *SnapshotFileReader) ReadAll() (data []byte, err error) {
	data, err = os.ReadFile(s.snapshotFile.Name())
	s.snapshotFile.Close()
	return
}

func (s *SnapshotFileReader) close() {
	s.snapshotFile.Close()
}

type SnapshotWriterInternal interface {
	SnapshotWriter
	close()
}

type SnapshotFileWriter struct {
	SnapshotFileInfo
}

func newSnapshotFileWriter(dir string, term uint32, logIndex uint64) (SnapshotWriter, error) {
	fileName := fmt.Sprintf("raft-snapshot_%d_%d.snapshot", term, logIndex)
	snapshotFile, err := os.Create(path.Join(dir, fileName))
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot file %s", err)
	}
	return &SnapshotFileWriter{
		SnapshotFileInfo{
			term:         term,
			logIndex:     logIndex,
			snapshotFile: snapshotFile,
		},
	}, nil
}

func (sw *SnapshotFileWriter) Write(data []byte) error {
	_, err := sw.snapshotFile.Write(data)
	return err
}

func (sw *SnapshotFileWriter) WriteAt(data []byte, offset int64) error {
	_, err := sw.snapshotFile.WriteAt(data, offset)
	return err
}

func (sw *SnapshotFileWriter) close() {
	sw.snapshotFile.Close()
}

type SnapshotManager struct {
	snapshotDir   string
	writeSupplier func(dir string, term uint32, logIndex uint64) (SnapshotWriter, error)
	readSupplier  func(dir string, term uint32, logIndex uint64) (SnapshotReader, error)
}

func newSnapshotManager(snapshotDir string) *SnapshotManager {
	err := os.MkdirAll(snapshotDir, 0755)
	if err != nil {
		logger.Fatal("Failed to create snapshot dir!", err)
	}
	return &SnapshotManager{
		snapshotDir:   snapshotDir,
		writeSupplier: newSnapshotFileWriter,
		readSupplier:  newSnapshotFileReader,
	}
}

func (s *SnapshotManager) CreateWriter(term uint32, logIndex uint64) (SnapshotWriter, error) {
	return s.writeSupplier(s.snapshotDir, term, logIndex)
}

func (*SnapshotManager) Close(writer SnapshotWriter) {
	writer.(SnapshotWriterInternal).close()
}

func (s *SnapshotManager) GetSnapshot(term uint32, logIndex uint64) (SnapshotReader, error) {
	return s.readSupplier(s.snapshotDir, term, logIndex)
}
