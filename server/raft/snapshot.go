package raft

import (
	"errors"
	"fmt"
	"github.com/pradeesh-kumar/go-raft-kv/logger"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"
)

const snapshotExtension = ".snapshot"

type SnapshotFileReader struct {
	snapshotFile *os.File
}

func newSnapshotFileReader(snapshotFile *os.File) (SnapshotReader, error) {
	return &SnapshotFileReader{snapshotFile}, nil
}

func (s *SnapshotFileReader) ReadAll() (data []byte, err error) {
	data, err = os.ReadFile(s.snapshotFile.Name())
	s.snapshotFile.Close()
	return
}

func (s *SnapshotFileReader) close() {
	s.snapshotFile.Close()
}

type SnapshotFileWriter struct {
	snapshotFile *os.File
}

func newSnapshotFileWriter(snapshotFile *os.File) (SnapshotWriter, error) {
	return &SnapshotFileWriter{snapshotFile}, nil
}

func (sw *SnapshotFileWriter) Write(data []byte) error {
	_, err := sw.snapshotFile.Write(data)
	return err
}

func (sw *SnapshotFileWriter) WriteAt(data []byte, offset int64) error {
	_, err := sw.snapshotFile.WriteAt(data, offset)
	return err
}

func (sw *SnapshotFileWriter) Close() {
	sw.snapshotFile.Close()
}

type FileBasedSnapshotManager struct {
	snapshotDir string
}

func newFileBasedSnapshotManager(snapshotDir string) SnapshotManager {
	err := os.MkdirAll(snapshotDir, 0755)
	if err != nil {
		logger.Fatal("Failed to create snapshot dir!", err)
	}
	return &FileBasedSnapshotManager{snapshotDir: snapshotDir}
}

func (s *FileBasedSnapshotManager) CreateWriter(term uint32, logIndex uint64) (SnapshotWriter, error) {
	fileName := fmt.Sprintf("raft-snapshot_%d_%d%s", term, logIndex, snapshotExtension)
	file, err := os.Open(path.Join(s.snapshotDir, fileName))
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("the snapshot file %s doesn't exists", fileName)
	}
	return newSnapshotFileWriter(file)
}

func (s *FileBasedSnapshotManager) GetSnapshot(term uint32, logIndex uint64) (SnapshotReader, error) {
	fileName := fmt.Sprintf("raft-snapshot_%d_%d%s", term, logIndex, snapshotExtension)
	file, err := os.Open(path.Join(s.snapshotDir, fileName))
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("the snapshot file %s doesn't exists", fileName)
	}
	return newSnapshotFileReader(file)
}

func (s *FileBasedSnapshotManager) GetLatestSnapshot() (SnapshotReader, error) {
	file, err := getLatestFile(s.snapshotDir, ".snapshot")
	if err != nil || file == nil {
		return nil, os.ErrNotExist
	}
	return newSnapshotFileReader(file)
}

func getLatestFile(dirPath string, ext string) (*os.File, error) {
	var latestFile string
	var latestModTime time.Time

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(d.Name()) == ext {
			info, err := d.Info()
			if err != nil {
				return err
			}
			if info.ModTime().After(latestModTime) {
				latestModTime = info.ModTime()
				latestFile = path
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if latestFile == "" {
		return nil, nil
	}
	return os.Open(latestFile)
}
