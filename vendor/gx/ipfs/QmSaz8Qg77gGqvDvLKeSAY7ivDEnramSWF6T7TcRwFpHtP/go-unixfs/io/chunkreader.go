package io

import (
	"context"
	"fmt"
	"io"

	mdag "gx/ipfs/QmRy4Qk9hbgFX9NGJRm8rBThrA8PZhNCitMgeRYyZ67s59/go-merkledag"
	ft "gx/ipfs/QmSaz8Qg77gGqvDvLKeSAY7ivDEnramSWF6T7TcRwFpHtP/go-unixfs"
	ftpb "gx/ipfs/QmSaz8Qg77gGqvDvLKeSAY7ivDEnramSWF6T7TcRwFpHtP/go-unixfs/pb"

	cid "gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	ipld "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
)

func prepend(b byte, buff []byte) []byte {
	return append([]byte{b}, buff...)
}

type ChunkReader struct {
	buff []byte
	dag  *chunkDagReader
}

func (cr *ChunkReader) ReadChunk() ([]byte, error) {
	if cr.buff != nil {
		out := cr.buff
		cr.buff = nil
		return out, nil
	} else if cr.dag != nil {
		return cr.dag.ReadChunk()
	}
	return nil, io.EOF
}

func (cr *ChunkReader) Close() error {
	cr.buff, cr.dag = nil, nil
	return nil
}

func NewChunkReader(ctx context.Context, n ipld.Node, serv ipld.NodeGetter) (*ChunkReader, error) {
	switch n := n.(type) {
	case *mdag.RawNode:
		return &ChunkReader{buff: prepend(0, n.RawData())}, nil
	case *mdag.ProtoNode:
		fsNode, err := ft.FSNodeFromBytes(n.Data())
		if err != nil {
			return nil, err
		}

		switch fsNode.Type() {
		case ftpb.Data_Directory, ftpb.Data_HAMTShard:
			// Dont allow reading directories
			return nil, ErrIsDir
		case ftpb.Data_File, ftpb.Data_Raw:
			return &ChunkReader{dag: newChunkDagReader(ctx, n, fsNode, serv)}, nil
		case ftpb.Data_Symlink:
			return nil, ErrCantReadSymlinks
		default:
			return nil, ft.ErrUnrecognizedType
		}
	default:
		return nil, ErrUnkownNodeType
	}
}

type chunkDagReader struct {
	serv ipld.NodeGetter

	// UnixFS file (it should be of type `Data_File` or `Data_Raw` only).
	file *ft.FSNode

	// the current data buffer to be read from
	buf *ChunkReader

	// NodePromises for each of 'nodes' child links
	promises []*ipld.NodePromise

	// the cid of each child of the current node
	links []*cid.Cid

	// the index of the child link currently being read from
	linkPosition int

	// current offset for the read head within the 'file'
	offset int64

	// Our context
	ctx context.Context

	// context cancel for children
	cancel func()
}

func newChunkDagReader(ctx context.Context, n *mdag.ProtoNode, file *ft.FSNode, serv ipld.NodeGetter) *chunkDagReader {
	fctx, cancel := context.WithCancel(ctx)
	curLinks := getLinkCids(n)
	return &chunkDagReader{
		serv:     serv,
		buf:      &ChunkReader{buff: prepend(1, n.RawData())},
		promises: make([]*ipld.NodePromise, len(curLinks)),
		links:    curLinks,
		ctx:      fctx,
		cancel:   cancel,
		file:     file,
	}
}

func (dr *chunkDagReader) preload(ctx context.Context, beg int) {
	end := beg + preloadSize
	if end >= len(dr.links) {
		end = len(dr.links)
	}

	copy(dr.promises[beg:], ipld.GetNodes(ctx, dr.serv, dr.links[beg:end]))
}

// precalcNextBuf follows the next link in line and loads it from the
// DAGService, setting the next buffer to read from
func (dr *chunkDagReader) precalcNextBuf(ctx context.Context) error {
	if dr.buf != nil {
		dr.buf.Close() // Just to make sure
		dr.buf = nil
	}

	if dr.linkPosition >= len(dr.promises) {
		return io.EOF
	}

	// If we drop to <= preloadSize/2 preloading nodes, preload the next 10.
	for i := dr.linkPosition; i < dr.linkPosition+preloadSize/2 && i < len(dr.promises); i++ {
		// TODO: check if canceled.
		if dr.promises[i] == nil {
			dr.preload(ctx, i)
			break
		}
	}

	nxt, err := dr.promises[dr.linkPosition].Get(ctx)
	dr.promises[dr.linkPosition] = nil
	switch err {
	case nil:
	case context.DeadlineExceeded, context.Canceled:
		err = ctx.Err()
		if err != nil {
			return ctx.Err()
		}
		// In this case, the context used to *preload* the node has been canceled.
		// We need to retry the load with our context and we might as
		// well preload some extra nodes while we're at it.
		//
		// Note: When using `Read`, this code will never execute as
		// `Read` will use the global context. It only runs if the user
		// explicitly reads with a custom context (e.g., by calling
		// `CtxReadFull`).
		dr.preload(ctx, dr.linkPosition)
		nxt, err = dr.promises[dr.linkPosition].Get(ctx)
		dr.promises[dr.linkPosition] = nil
		if err != nil {
			return err
		}
	default:
		return err
	}

	dr.linkPosition++

	return dr.loadBufNode(nxt)
}

func (dr *chunkDagReader) loadBufNode(node ipld.Node) error {
	switch node := node.(type) {
	case *mdag.ProtoNode:
		fsNode, err := ft.FSNodeFromBytes(node.Data())
		if err != nil {
			return fmt.Errorf("incorrectly formatted protobuf: %s", err)
		}

		switch fsNode.Type() {
		case ftpb.Data_File:
			dr.buf = &ChunkReader{dag: newChunkDagReader(dr.ctx, node, fsNode, dr.serv)}
			return nil
		case ftpb.Data_Raw:
			dr.buf = &ChunkReader{buff: prepend(1, node.RawData())}
			return nil
		default:
			return fmt.Errorf("found %s node in unexpected place", fsNode.Type().String())
		}
	case *mdag.RawNode:
		dr.buf = &ChunkReader{buff: prepend(0, node.RawData())}
		return nil
	default:
		return ErrUnkownNodeType
	}
}

// Read reads data from the DAG structured file
func (dr *chunkDagReader) ReadChunk() ([]byte, error) {
	return dr.ctxReadChunk(dr.ctx)
}

// CtxReadFull reads data from the DAG structured file
func (dr *chunkDagReader) ctxReadChunk(ctx context.Context) ([]byte, error) {
	if dr.buf == nil {
		if err := dr.precalcNextBuf(ctx); err != nil {
			return nil, err
		}
	}

	for {
		chunk, err := dr.buf.ReadChunk()
		if err == nil {
			return chunk, nil
		} else if err != io.EOF {
			return nil, err
		}

		// if we are not done with the output buffer load next block
		err = dr.precalcNextBuf(ctx)
		if err != nil {
			return nil, err
		}
	}
}

// Close closes the reader.
func (dr *chunkDagReader) Close() error {
	dr.cancel()
	return nil
}
