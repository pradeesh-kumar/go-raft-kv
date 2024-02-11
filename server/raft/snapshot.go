package raft

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/pradeesh-kumar/go-raft-kv/logger"
	"google.golang.org/protobuf/proto"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// .snapshot file format
// First 64bit -> offset of the snapshot data.
// The config entry starts after the first 64bit till the defined offset.
// From the defined offset, snapshot data will be present till the end of the file.

const snapshotExtension = ".snapshot"

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

func (s *FileBasedSnapshotManager) CreateWriter(term uint32, logIndex uint64, configEntry *ConfigEntry) (SnapshotWriter, error) {
	fileName := fmt.Sprintf("raft-snapshot_%d_%d%s", term, logIndex, snapshotExtension)
	file, err := os.OpenFile(path.Join(s.snapshotDir, fileName), os.O_CREATE|os.O_WRONLY, 0666)
	if errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("the snapshot file %s already exists", fileName)
	}
	b, err := proto.Marshal(configEntry)
	if err != nil {
		return nil, errors.New("failed to marshall the config entry to create the snapshot writer")
	}
	// Write the config offset to the snapshot
	configOffset := uint64(8 + len(b))
	offsetBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(offsetBytes, configOffset)
	_, err = file.Write(offsetBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to write the config offset %d to the snapshot file", configOffset)
	}

	// Write config entry to the snapshot
	_, err = file.Write(b)
	if err != nil {
		return nil, errors.New("failed to write config entry to the snapshot file")
	}
	return file, nil
}

func (s *FileBasedSnapshotManager) GetSnapshot(term uint32, logIndex uint64) (*Snapshot, error) {
	fileName := fmt.Sprintf("raft-snapshot_%d_%d%s", term, logIndex, snapshotExtension)
	file, err := os.Open(path.Join(s.snapshotDir, fileName))
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("the snapshot file %s doesn't exists", fileName)
	}
	return getSnapshotFromFile(file)
}

func (s *FileBasedSnapshotManager) GetLatestSnapshot() (*Snapshot, error) {
	file, err := getLatestFile(s.snapshotDir, ".snapshot")
	if err != nil || file == nil {
		return nil, os.ErrNotExist
	}
	return getSnapshotFromFile(file)
}

func getSnapshotFromFile(file *os.File) (*Snapshot, error) {
	snapshot := &Snapshot{}
	// Read the 64bit offset where the config entry ends
	b := make([]byte, 8)
	n, err := file.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to read the snapshot file: %s", err)
	}
	if n == 0 {
		return nil, os.ErrNotExist
	}
	configEntryEndOffset := binary.LittleEndian.Uint64(b)

	// Read config entry
	b = make([]byte, configEntryEndOffset-8)
	n, err = file.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to read the snapshot file: %s", err)
	}
	if n < int(configEntryEndOffset) {
		return nil, errors.New("snapshot doesn't contain the config entry")
	}
	snapshot.configEntry = &ConfigEntry{}
	err = proto.Unmarshal(b, snapshot.configEntry)
	if err != nil {
		return nil, errors.New("config entry corrupted in the snapshot")
	}
	snapshot.snapshotReader = file
	snapshot.lastLogTerm, snapshot.lastLogIndex = parseTermAndIndex(file.Name())
	return snapshot, nil
}

func parseTermAndIndex(fileName string) (term uint32, index uint64) {
	fileName = fileName[strings.Index(fileName, "_")+1:]
	converted, _ := strconv.ParseUint(fileName[:strings.Index(fileName, "_")], 10, 32)
	term = uint32(converted)

	fileName = fileName[strings.Index(fileName, "_")+1:]
	index, _ = strconv.ParseUint(fileName[:strings.Index(fileName, ".")], 10, 64)
	return
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
