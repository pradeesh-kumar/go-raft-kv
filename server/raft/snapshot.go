package raft

import (
	"fmt"
	"github.com/pradeesh-kumar/go-raft-kv/logger"
	"os"
	"path"
)

type SnapshotWriterInternal interface {
	SnapshotWriter
	close()
}

type FileSnapshotWriter struct {
	term         uint32
	logIndex     uint64
	snapshotFile *os.File
}

func newFileSnapshotWriter(dir string, term uint32, logIndex uint64) (SnapshotWriter, error) {
	fileName := fmt.Sprintf("raft-snapshot_%d_%d.snapshot", term, logIndex)
	snapshotFile, err := os.Create(path.Join(dir, fileName))
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot file %s", err)
	}
	return &FileSnapshotWriter{
		term:         term,
		logIndex:     logIndex,
		snapshotFile: snapshotFile,
	}, nil
}

func (sw *FileSnapshotWriter) WriteBytes(data []byte) error {
	_, err := sw.snapshotFile.Write(data)
	return err
}

func (sw *FileSnapshotWriter) close() {
	sw.snapshotFile.Close()
}

type SnapshotManager struct {
	snapshotDir          string
	createWriterSupplier func(dir string, term uint32, logIndex uint64) (SnapshotWriter, error)
}

func newSnapshotManager(snapshotDir string) *SnapshotManager {
	err := os.MkdirAll(snapshotDir, 0755)
	if err != nil {
		logger.Fatal("Failed to create snapshot dir!", err)
	}
	return &SnapshotManager{
		snapshotDir:          snapshotDir,
		createWriterSupplier: newFileSnapshotWriter,
	}
}

func (s *SnapshotManager) CreateWriter(term uint32, logIndex uint64) (SnapshotWriter, error) {
	return s.createWriterSupplier(s.snapshotDir, term, logIndex)
}

func (*SnapshotManager) Capture(writer SnapshotWriter) {
	writer.(SnapshotWriterInternal).close()
}
