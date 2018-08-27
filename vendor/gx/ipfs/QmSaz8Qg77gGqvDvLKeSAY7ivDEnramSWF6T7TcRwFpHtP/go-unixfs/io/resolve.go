package io

import (
	"context"
	"fmt"

	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"

	dag "gx/ipfs/QmRy4Qk9hbgFX9NGJRm8rBThrA8PZhNCitMgeRYyZ67s59/go-merkledag"
	ft "gx/ipfs/QmSaz8Qg77gGqvDvLKeSAY7ivDEnramSWF6T7TcRwFpHtP/go-unixfs"
	hamt "gx/ipfs/QmSaz8Qg77gGqvDvLKeSAY7ivDEnramSWF6T7TcRwFpHtP/go-unixfs/hamt"
	resolver "gx/ipfs/QmYKNMEUK7nCVAefgXF1LVtZEZg3uRmBqiae4FJRXDNAyJ/go-path/resolver"

	ipld "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
)

// ResolveUnixfsOnce resolves a single hop of a path through a graph in a
// unixfs context. This includes handling traversing sharded directories.
func ResolveUnixfsOnce(ctx context.Context, ds ipld.NodeGetter, nd ipld.Node, names []string) (*ipld.Link, []string, error) {
	switch nd := nd.(type) {
	case *dag.ProtoNode:
		upb, err := ft.FromBytes(nd.Data())
		if err != nil {
			// Not a unixfs node, use standard object traversal code
			lnk, err := nd.GetNodeLink(names[0])
			if err != nil {
				return nil, nil, err
			}

			return lnk, names[1:], nil
		}

		switch upb.GetType() {
		case ft.THAMTShard:
			rods := dag.NewReadOnlyDagService(ds)
			s, err := hamt.NewHamtFromDag(rods, nd)
			if err != nil {
				return nil, nil, err
			}

			out, err := s.Find(ctx, names[0])
			if err != nil {
				return nil, nil, err
			}

			return out, names[1:], nil
		default:
			lnk, err := nd.GetNodeLink(names[0])
			if err != nil {
				return nil, nil, err
			}

			return lnk, names[1:], nil
		}
	default:
		lnk, rest, err := nd.ResolveLink(names)
		if err != nil {
			return nil, nil, err
		}
		return lnk, rest, nil
	}
}

func SecureResolveUnixfsOnce(cw coreiface.ChunkWriter) resolver.ResolveOnce {
	return func(ctx context.Context, ds ipld.NodeGetter, nd ipld.Node, names []string) (*ipld.Link, []string, error) {
		switch nd := nd.(type) {
		case *dag.ProtoNode:
			upb, err := ft.FromBytes(nd.Data())
			if err != nil {
				// Not a unixfs node, use standard object traversal code
				lnk, err := nd.GetNodeLink(names[0])
				if err != nil {
					return nil, nil, err
				}
				if err := cw.WriteChunk(prepend(0, nd.RawData())); err != nil {
					return nil, nil, err
				}
				return lnk, names[1:], nil
			}

			switch upb.GetType() {
			case ft.THAMTShard:
				rods := dag.NewReadOnlyDagService(ds)
				s, err := hamt.NewHamtFromDag(rods, nd)
				if err != nil {
					return nil, nil, err
				}

				out, err := s.SecureFind(ctx, cw, names[0])
				if err != nil {
					return nil, nil, err
				}

				return out, names[1:], nil
			default:
				lnk, err := nd.GetNodeLink(names[0])
				if err != nil {
					return nil, nil, err
				}
				if err := cw.WriteChunk(prepend(0, nd.RawData())); err != nil {
					return nil, nil, err
				}
				return lnk, names[1:], nil
			}
		default:
			return nil, nil, fmt.Errorf("resolving through generic node is not implemented")
		}
	}
}