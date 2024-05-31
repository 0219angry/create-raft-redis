package transport

import (
	"log"
	"net"

	"github.com/bootjp/go-kvlib/store"
	hraft "github.com/hashicorp/raft"
	"github.com/tidwall/redcon"
)

type Redis struct {
	listen      net.Listener
	store       store.Store
	stableStore hraft.StableStore
	id          hraft.ServerID
	raft        *hraft.Raft
}

/* ==============================================================================
 *    Constructor: Create a new Redis transport
 * ============================================================================*/
func NewRedis(id hraft.ServerID, raft *hraft.Raft, store store.Store, stableStore hraft.StableStore) *Redis {
	return &Redis{
		store:       store,
		raft:        raft,
		id:          id,
		stableStore: stableStore,
	}
}

/* ==============================================================================
 *
 * ============================================================================*/
func (r *Redis) Serve(addr string) error {
	var err error
	r.listen, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return r.handle()
}

func (r *Redis) handle() error {
	return redcon.Serve(r.listen,
		func(conn redcon.Conn, cmd redcon.Command) {
			err := r.validateCmd(cmd)
			if err != nil {
				conn.WriteError(err.Error())
				return
			}
			r.processCmd(conn, cmd)
		},
		func(conn redcon.Conn) bool {
			return true
		},
		func(conn redcon.Conn, err error) {
			if err != nil {
				log.Default().Println("error:", err)
			}
		},
	)
}

var argLens = map[string]int{
	"GET": 2,
	"SET": 3,
	"DEL": 2,
}

/* ==============================================================================
 *    Validate command arguments length
 * ============================================================================*/
func (r *Redis) validateCmd(cmd redcon.Command) error {

}

func (r *Redis) processCmd(conn redcon.Conn, cmd redcon.Command) {

}
