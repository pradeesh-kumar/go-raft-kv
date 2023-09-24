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
	return &DeleteCommand{
		KVCommandProperty: KVCommandProperty{
			key: parseKey(cmd),
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
	return &InsertCommand{
		KVCommandProperty: KVCommandProperty{
			key: parseKey(cmd),
			val: parseVal(cmd),
			cmd: cmd,
		},
	}
}

func parseKey(cmd string) string {
	begIdx := strings.Index(cmd, "KEY=") + 4
	cmd = cmd[begIdx:]
	endIdx := strings.Index(cmd, " ")
	if endIdx != -1 {
		return cmd[:endIdx]
	}
	return cmd
}

func parseVal(cmd string) string {
	begIdx := strings.Index(cmd, "VALUE=") + 6
	return cmd[begIdx:]
}
