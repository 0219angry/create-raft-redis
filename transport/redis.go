package transport

import (
	"context"
	"errors"
	"log"
	"net"
	"strings"

	"create-raft-redis/store"

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

var argsLen = map[string]int{
	"GET": 2,
	"SET": 3,
	"DEL": 2,
}

const (
	commandName = 0
	keyName     = 1
	value       = 2
)

/* ==============================================================================
 *    Validate command arguments length
 * ============================================================================*/
func (r *Redis) validateCmd(cmd redcon.Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "'command")
	}
	if len(cmd.Args) < argsLen[string(cmd.Args[commandName])] {
		return errors.New("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "'command")
	}

	/* Check args length */
	plainCmd := strings.ToUpper(string(cmd.Args[commandName]))

	if len(cmd.Args) != argsLen[plainCmd] {
		return errors.New("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "'command")
	}
	return nil

}

func (r *Redis) processCmd(conn redcon.Conn, cmd redcon.Command) {
	ctx := context.Background()
	/*
	 *	check if the node is the leader
	 */
	if r.raft.State() != hraft.Leader {
		_, lid := r.raft.LeaderWithID()
		add, err := store.GetRedisAddrByNodeID(r.stableStore, lid)
		if err != nil {
			conn.WriteError(err.Error())
			return
		}

		conn.WriteError("MOVED -1 " + add)
		return
	}

	planCmd := strings.ToUpper(string(cmd.Args[commandName]))
	switch planCmd {
	case "GET":
		val, err := r.store.Get(ctx, cmd.Args[keyName])
		if err != nil {
			switch {
			case errors.Is(err, store.ErrKeyNotFound):
				conn.WriteNull()
				return
			default:
				conn.WriteError(err.Error())
				return
			}
		}
		conn.WriteBulk(val)
	}
}
