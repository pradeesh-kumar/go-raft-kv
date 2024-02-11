package raft

import (
	"strings"
	"testing"
)

func Test_parseTermAndIndex(t *testing.T) {
	fileName := "raft-snapshot_123_222.snapshot"
	println(fileName)
	fileName = fileName[strings.Index(fileName, "_")+1:]
	println(fileName)
	term := fileName[:strings.Index(fileName, "_")]
	println(term)

	fileName = fileName[strings.Index(fileName, "_")+1:]
	logIndex := fileName[:strings.Index(fileName, ".")]
	println(logIndex)
}
