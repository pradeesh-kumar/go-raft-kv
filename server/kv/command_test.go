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

	deserialized := NewInsertCommandFromRaw(insertCommand.cmd)
	assert.Equal(t, "name", deserialized.key)
	assert.Equal(t, "raft", deserialized.val)
}

func TestParseDeleteCommand(t *testing.T) {
	key := "hello"
	deleteCommand := NewDeleteCommandFromKey(key)
	assert.Equal(t, fmt.Sprintf("DELETE KEY=%s", key), deleteCommand.cmd)

	deserialized := NewDeleteCommandFromRaw(deleteCommand.cmd)
	assert.Equal(t, key, deserialized.key)
}
