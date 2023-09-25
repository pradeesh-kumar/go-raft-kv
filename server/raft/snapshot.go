package raft

import "os"

type SnapshotWriterInternal interface {
	SnapshotWriter
	close()
}

type FileSnapshotWriter struct {
	snapshotFile *os.File
}

func (*FileSnapshotWriter) WriteBytes(data []byte) {
}

func (*FileSnapshotWriter) close() {
}

func newFileSnapshotWriter() *FileSnapshotWriter {
	return &FileSnapshotWriter{}
}

type FileBasedSnapshotManager struct {
}

func newFileBasedSnapshotManager() *FileBasedSnapshotManager {
	return &FileBasedSnapshotManager{}
}

func (*FileBasedSnapshotManager) CreateWriter() SnapshotWriter {
	return newFileSnapshotWriter()
}

func (*FileBasedSnapshotManager) Capture(writer SnapshotWriter) {
	writer.(SnapshotWriterInternal).close()
}
