package rlog

import (
	"github.com/pradeesh-kumar/go-raft-kv/raft"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
)

func TestSegment(t *testing.T) {
	dir, _ := os.MkdirTemp("", "segment-test")
	defer os.RemoveAll(dir)

	want := &raft.Record{
		LogEntryBody: &raft.Record_StateMachineEntry{
			StateMachineEntry: &raft.StateMachineEntry{Value: []byte("hello world")},
		},
	}

	c := LogConfig{}
	c.MaxStoreBytes = 1024
	c.MaxIndexBytes = entryWidth * 3

	s, err := newSegment(dir, 16, c)
	require.NoError(t, err)
	require.Equal(t, uint64(16), s.nextOffset, s.nextOffset)
	require.False(t, s.IsMaxed())

	for i := uint64(0); i < 3; i++ {
		off, err := s.Append(want)
		require.NoError(t, err)
		require.Equal(t, 16+i, off)

		got, err := s.Read(off)
		require.NoError(t, err)
		require.Equal(t, want.LogEntryBody.(*raft.Record_StateMachineEntry).StateMachineEntry.GetValue(), got.LogEntryBody.(*raft.Record_StateMachineEntry).StateMachineEntry.GetValue())
	}

	_, err = s.Append(want)
	require.Equal(t, io.EOF, err)

	// maxed index
	require.True(t, s.IsMaxed())

	c.MaxStoreBytes = uint64(len(want.LogEntryBody.(*raft.Record_StateMachineEntry).StateMachineEntry.GetValue()) * 3)
	c.MaxIndexBytes = 1024

	s, err = newSegment(dir, 16, c)
	require.NoError(t, err)
	// maxed store
	require.True(t, s.IsMaxed())

	err = s.Remove()
	require.NoError(t, err)
	s, err = newSegment(dir, 16, c)
	require.NoError(t, err)
	require.False(t, s.IsMaxed())
}
