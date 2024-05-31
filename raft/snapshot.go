package raft

import (
	"io"

	"github.com/hashicorp/raft"
)

var _ raft.FSMSnapshot = (*KVSnapshot)(nil)

type KVSnapshot struct {
	io.ReadWriter
}

/* ==============================================================================
 *    Snapshot取得時に呼び出される
 *	  取得したSnapshotをsinkに書き込む
 * ============================================================================*/
func (f *KVSnapshot) Persist(sink raft.SnapshotSink) error {
	defer sink.Close()
	_, err := io.Copy(sink, f)
	return err
}

func (f *KVSnapshot) Release() {
}
