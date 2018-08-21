package io

import (
	"io"
)

type ChunkBuffer struct {
	buff [][]byte
}

func (cb *ChunkBuffer) WriteChunk(data []byte) error {
	cb.buff = append(cb.buff, data)
	return nil
}

func (cb *ChunkBuffer) ReadChunk() ([]byte, error) {
	if len(cb.buff) == 0 {
		return nil, io.EOF
	}
	out := cb.buff[0]
	cb.buff = cb.buff[1:]
	return out, nil
}

func (cb *ChunkBuffer) Close() error {
	cb.buff = nil
	return nil
}
