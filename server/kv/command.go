package kv

import (
	"fmt"
	"strings"
)

func parseCommand(bytes []byte) (KVCommand, error) {
	cmd := string(bytes)
	if strings.HasPrefix(cmd, "INSERT") {
		return NewInsertCommandFromRaw(cmd), nil
	} else if strings.HasPrefix(cmd, "DELETE") {
		return NewDeleteCommandFromRaw(cmd), nil
	} else {
		return nil, fmt.Errorf("unrecognized command %s", cmd)
	}
}

type KVCommand interface {
}

type KVCommandProperty struct {
	key string
	val string
	cmd string
}

type DeleteCommand struct {
	KVCommand
	KVCommandProperty
}

func NewDeleteCommandFromKey(key string) *DeleteCommand {
	return &DeleteCommand{
		KVCommandProperty: KVCommandProperty{
			key: key,
			cmd: fmt.Sprintf("DELETE KEY=%s", key),
		},
	}
}

func NewDeleteCommandFromRaw(cmd string) *DeleteCommand {
	after, _ := strings.CutPrefix(cmd, "DELETE KEY=")
	keyIndex := strings.Index(after, "")
	key := after[:keyIndex]
	return &DeleteCommand{
		KVCommandProperty: KVCommandProperty{
			key: key,
			cmd: cmd,
		},
	}
}

func (c *DeleteCommand) Bytes() []byte {
	return []byte(c.cmd)
}

type InsertCommand struct {
	KVCommand
	KVCommandProperty
}

func (c *InsertCommand) Bytes() []byte {
	return []byte(c.cmd)
}

func NewInsertCommandFromKV(key string, val string) *InsertCommand {
	return &InsertCommand{
		KVCommandProperty: KVCommandProperty{
			key: key,
			val: val,
			cmd: fmt.Sprintf("INSERT KEY=%s VALUE=%s", key, val),
		},
	}
}

func NewInsertCommandFromRaw(cmd string) *InsertCommand {
	after, _ := strings.CutPrefix(cmd, "INSERT KEY=")
	keyIndex := strings.Index(after, "")
	key := after[:keyIndex]

	valIndex := strings.Index(after, "=")
	val := after[valIndex:]
	return &InsertCommand{
		KVCommandProperty: KVCommandProperty{
			key: key,
			val: val,
			cmd: cmd,
		},
	}
}
