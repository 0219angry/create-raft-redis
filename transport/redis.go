package transport

import (
	"net"

	hraft "github.com/hashicorp/raft"
)

type redis struct {
	listen      net.Listener
	store       store.Store
	stableStore hraft.StableStore
}
