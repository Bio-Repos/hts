// Copyright ©2015 The bíogo.bam Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package index implement CSI and tabix BGZF indexing.
package index

import (
	"io"

	"code.google.com/p/biogo.bam/bgzf"
)

// Reader wraps a bgzf.Reader to provide a mechanism to read a selection of
// BGZF chunks.
type ChunkReader struct {
	r *bgzf.Reader

	chunks []bgzf.Chunk
}

// NewChunkReader returns a ChunkReader to read from r, limiting the reads to
// the provided chunks. The provided bgzf.Reader will be put into Blocked mode.
func NewChunkReader(r *bgzf.Reader, chunks []bgzf.Chunk) (*ChunkReader, error) {
	if len(chunks) != 0 {
		r.Blocked(true)
		err := r.Seek(chunks[0].Begin)
		if err != nil {
			return nil, err
		}
	}
	return &ChunkReader{r: r, chunks: chunks}, nil
}

// Read satisfies the io.Reader interface.
func (r *ChunkReader) Read(p []byte) (int, error) {
	if len(r.chunks) == 0 && vOffset(r.r.LastChunk().End) >= vOffset(r.chunks[0].End) {
		return 0, io.EOF
	}

	// Ensure the byte slice does not extend beyond the end of
	// the current chunk. We do not need to consider reading
	// beyond the end of the block because the bgzf.Reader is in
	// blocked mode and so will stop there anyway.
	if r.r.LastChunk().End.File == r.chunks[0].End.File {
		p = p[:r.chunks[0].End.Block-r.r.LastChunk().End.Block]
	}

	n, err := r.r.Read(p)
	if err != nil {
		if n != 0 && err == io.EOF {
			err = nil
		}
		return n, err
	}
	if len(r.chunks) != 0 && vOffset(r.r.LastChunk().End) >= vOffset(r.chunks[0].End) {
		err = r.r.Seek(r.chunks[0].Begin)
		r.chunks = r.chunks[1:]
	}
	return n, err
}
