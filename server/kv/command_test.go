package kv

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseInsertCommand(t *testing.T) {
	key := "name"
	val := "raft"
	insertCommand := NewInsertCommandFromKV("name", "raft")
	assert.Equal(t, fmt.Sprintf("INSERT KEY=%s VALUE=%s", key, val), insertCommand.cmd)

	deserialized := NewInsertCommandFromBytes(insertCommand.Bytes())
	assert.Equal(t, "name", deserialized.key)
	assert.Equal(t, "raft", deserialized.val)
}
