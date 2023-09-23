package kv

import (
	"fmt"
	"strings"
)

type InsertCommand struct {
	key string
	val string
	cmd string
}

func (c *InsertCommand) Bytes() []byte {
	return []byte(c.cmd)
}

func NewInsertCommandFromKV(key string, val string) InsertCommand {
	return InsertCommand{
		key: key,
		val: val,
		cmd: fmt.Sprintf("INSERT KEY=%s VALUE=%s", key, val),
	}
}

func NewInsertCommandFromBytes(bytes []byte) InsertCommand {
	cmd := string(bytes)
	return parseInsertCommand(cmd)
}

func parseInsertCommand(cmd string) InsertCommand {
	after, _ := strings.CutPrefix(cmd, "INSERT KEY=")
	keyIndex := strings.Index(after, "")
	key := after[:keyIndex]

	valIndex := strings.Index(after, "=")
	val := after[valIndex:]
	return InsertCommand{key, val, cmd}
}
