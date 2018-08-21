package iface

import (
	"context"
	"io"

	ipld "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
)

// UnixfsAPI is the basic interface to immutable files in IPFS
type UnixfsAPI interface {
	// Add imports the data from the reader into merkledag file
	Add(context.Context, io.Reader) (ResolvedPath, error)

	// Cat returns a reader for the file
	Cat(context.Context, Path) (Reader, error)

	CatChunks(context.Context, Path) (ChunkReader, error)

	// Ls returns the list of links in a directory
	Ls(context.Context, Path) ([]*ipld.Link, error)
}

type ChunkReader interface {
	ReadChunk() ([]byte, error)
	Close() error
}

type ChunkWriter interface {
	WriteChunk([]byte) error
}
