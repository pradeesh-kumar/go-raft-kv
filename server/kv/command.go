package kv

import "fmt"

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
	return InsertCommand{}
}
