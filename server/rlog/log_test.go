package rlog

import (
	"github.com/pradeesh-kumar/go-raft-kv/raft"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T, log *Log,
	){
		"append and read a record succeeds": testAppendRead,
		"offset out of range error":         testOutOfRangeErr,
		"init with existing segments":       testInitExisting,
		"reader":                            testReader,
		"truncate":                          testTruncate,
	} {
		t.Run(scenario, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "store-test")
			require.NoError(t, err)
			defer os.RemoveAll(dir)

			c := LogConfig{}
			c.MaxStoreBytes = 32
			log, err := NewLog(dir, c)
			require.NoError(t, err)

			fn(t, log)
		})
	}
}

func testAppendRead(t *testing.T, log *Log) {
	append := &raft.Record{
		LogEntryBody: &raft.Record_StateMachineEntry{
			StateMachineEntry: &raft.StateMachineEntry{Value: []byte("hello world")},
		},
	}
	off, err := log.Append(append)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	read, err := log.Read(off)
	require.NoError(t, err)
	require.Equal(t, append.LogEntryBody.(*raft.Record_StateMachineEntry).StateMachineEntry.GetValue(), read.LogEntryBody.(*raft.Record_StateMachineEntry).StateMachineEntry.GetValue())
}

func testOutOfRangeErr(t *testing.T, log *Log) {
	read, err := log.Read(1)
	require.Nil(t, read)
	require.Error(t, err)
}

func testInitExisting(t *testing.T, o *Log) {
	append := &raft.Record{
		LogEntryBody: &raft.Record_StateMachineEntry{
			StateMachineEntry: &raft.StateMachineEntry{Value: []byte("hello world")},
		},
	}
	for i := 0; i < 3; i++ {
		_, err := o.Append(append)
		require.NoError(t, err)
	}
	require.NoError(t, o.Close())

	off := o.LowestOffset()
	require.Equal(t, uint64(0), off)
	off = o.HighestOffset()
	require.Equal(t, uint64(2), off)

	n, err := NewLog(o.Dir, o.Config)
	require.NoError(t, err)

	off = n.LowestOffset()
	require.Equal(t, uint64(0), off)
	off = n.HighestOffset()
	require.Equal(t, uint64(2), off)
}

func testReader(t *testing.T, log *Log) {
	append := &raft.Record{
		LogEntryBody: &raft.Record_StateMachineEntry{
			StateMachineEntry: &raft.StateMachineEntry{Value: []byte("hello world")},
		},
	}
	off, err := log.Append(append)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	reader := log.Reader()
	b, err := ioutil.ReadAll(reader)
	require.NoError(t, err)

	read := &raft.Record{}
	err = proto.Unmarshal(b[lenWidth:], read)
	require.NoError(t, err)
	require.Equal(t, append.LogEntryBody.(*raft.Record_StateMachineEntry).StateMachineEntry.GetValue(), read.LogEntryBody.(*raft.Record_StateMachineEntry).StateMachineEntry.GetValue())
}

func testTruncate(t *testing.T, log *Log) {
	append := &raft.Record{
		LogEntryBody: &raft.Record_StateMachineEntry{
			StateMachineEntry: &raft.StateMachineEntry{Value: []byte("hello world")},
		},
	}
	for i := 0; i < 3; i++ {
		_, err := log.Append(append)
		require.NoError(t, err)
	}

	err := log.Truncate(1)
	require.NoError(t, err)

	_, err = log.Read(0)
	require.Error(t, err)
}
