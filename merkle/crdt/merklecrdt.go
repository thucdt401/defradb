// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package crdt

import (
	"context"

	"github.com/sourcenetwork/defradb/core"
	corenet "github.com/sourcenetwork/defradb/core/net"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
)

// var (
//     log = logging.Logger("defradb.merkle.crdt")
// )

// MerkleCRDT is the implementation of a Merkle Clock along with a
// CRDT payload. It implements the ReplicatedData interface
// so it can be merged with any given semantics.
type MerkleCRDT interface {
	core.ReplicatedData
	Clock() core.MerkleClock
}

// type MerkleCRDTInitFn func(ds.Key) MerkleCRDT
// type MerkleCRDTFactory func(store core.DSReaderWriter, namespace ds.Key) MerkleCRDTInitFn

// Type indicates MerkleCRDT type
// type Type byte

// const (
// 	//no lint
// 	none = Type(iota) // reserved none type
// 	LWW_REGISTER
// 	OBJECT
// )

var (
	// defaultMerkleCRDTs                     = make(map[Type]MerkleCRDTFactory)
	_ core.ReplicatedData = (*baseMerkleCRDT)(nil)
)

// The baseMerkleCRDT handles the merkle crdt overhead functions
// that aren't CRDT specific like the mutations and state retrieval
// functions. It handles creating and publishing the crdt DAG with
// the help of the MerkleClock
type baseMerkleCRDT struct {
	clock core.MerkleClock
	crdt  core.ReplicatedData

	broadcaster corenet.Broadcaster
}

func (base *baseMerkleCRDT) Clock() core.MerkleClock {
	return base.clock
}

func (base *baseMerkleCRDT) Merge(ctx context.Context, other core.Delta, id string) error {
	return base.crdt.Merge(ctx, other, id)
}

func (base *baseMerkleCRDT) DeltaDecode(node ipld.Node) (core.Delta, error) {
	return base.crdt.DeltaDecode(node)
}

func (base *baseMerkleCRDT) Value(ctx context.Context) ([]byte, error) {
	return base.crdt.Value(ctx)
}

func (base *baseMerkleCRDT) ID() string {
	return base.crdt.ID()
}

// Publishes the delta to state
func (base *baseMerkleCRDT) Publish(ctx context.Context, delta core.Delta, broadcast bool) (cid.Cid, error) {
	c, err := base.clock.AddDAGNode(ctx, delta)
	if err != nil {
		return cid.Undef, err
	}
	// and broadcast
	if base.broadcaster != nil && broadcast {
		go func() {
			log := core.Log{
				DocKey: base.crdt.ID(),
				CID:    c,
				Delta:  delta,
			}
			base.broadcaster.Broadcast(log) //@todo
		}()
	}
	return c, nil
}
