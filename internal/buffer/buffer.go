// Copyright 2014-2019 Ulrich Kunitz. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buffer

import (
	"errors"
)

// Buffer provides a circular Buffer of bytes. If the front index equals
// the rear index the Buffer is empty. As a consequence front cannot be
// equal rear for a full Buffer. So a full Buffer has a length that is
// one byte less the the length of the data slice.
type Buffer struct {
	data  []byte
	front int
	rear  int
}

type DecBuf interface {
	Available() int
	Buffered() int
	Cap() int
	Peek(p []byte) (n int, err error)
	PeekTail(dist int64, length int) ([]byte, error)
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	WriteByte(c byte) error
}

// newBuffer creates a buffer with the given size.
func NewBuffer(size int) *Buffer {
	return &Buffer{data: make([]byte, size+1)}
}

// Cap returns the capacity of the buffer.
func (b *Buffer) Cap() int {
	return len(b.data) - 1
}

// Resets the buffer. The front and rear index are set to zero.
func (b *Buffer) Reset() {
	b.front = 0
	b.rear = 0
}

// Buffered returns the number of byte buffered.
func (b *Buffer) Buffered() int {
	delta := b.front - b.rear
	if delta < 0 {
		delta += len(b.data)
	}
	return delta
}

// Available returns the number of bytes available for writing.
func (b *Buffer) Available() int {
	delta := b.rear - 1 - b.front
	if delta < 0 {
		delta += len(b.data)
	}
	return delta
}

// addIndex adds a non-negative integer to the index i and returns the
// resulting index. The function takes care of wrapping the index as
// well as potential overflow situations.
func (b *Buffer) addIndex(i int, n int) int {
	// subtraction of len(b.data) prevents overflow
	i += n - len(b.data)
	if i < 0 {
		i += len(b.data)
	}
	return i
}

// Read reads bytes from the buffer into p and returns the number of
// bytes read. The function never returns an error but might return less
// data than requested.
func (b *Buffer) Read(p []byte) (n int, err error) {
	n, err = b.Peek(p)
	b.rear = b.addIndex(b.rear, n)
	return n, err
}

// Peek reads bytes from the buffer into p without changing the buffer.
// Peek will never return an error but might return less data than
// requested.
func (b *Buffer) Peek(p []byte) (n int, err error) {
	m := b.Buffered()
	n = len(p)
	if m < n {
		n = m
		p = p[:n]
	}
	k := copy(p, b.data[b.rear:])
	if k < n {
		copy(p[k:], b.data)
	}
	return n, nil
}

// PeekTail returns a view into the buffer n bytes before its end with length at
// most the given length. It may return fewer than length bytes, but should not
// return zero bytes and nil error.
//
// If length > dist, at most dist bytes are returned, but this is not an error.
func (b *Buffer) PeekTail(dist int64, length int) ([]byte, error) {
	i := b.front - int(dist)
	if i < 0 {
		i += len(b.data)
	}

	var p []byte
	if i >= b.front {
		p = b.data[i:]
	} else {
		p = b.data[i:b.front]
	}

	if len(p) > length {
		p = p[:length]
	}

	return p, nil
}

// Discard skips the n next bytes to read from the buffer, returning the
// bytes discarded.
//
// If Discards skips fewer than n bytes, it returns an error.
func (b *Buffer) Discard(n int) (discarded int, err error) {
	if n < 0 {
		return 0, errors.New("buffer.Discard: negative argument")
	}
	m := b.Buffered()
	if m < n {
		n = m
		err = errors.New(
			"buffer.Discard: discarded less bytes then requested")
	}
	b.rear = b.addIndex(b.rear, n)
	return n, err
}

// ErrNoSpace indicates that there is insufficient space for the Write
// operation.
var ErrNoSpace = errors.New("insufficient space")

// Write puts data into the  buffer. If less bytes are written than
// requested ErrNoSpace is returned.
func (b *Buffer) Write(p []byte) (n int, err error) {
	m := b.Available()
	n = len(p)
	if m < n {
		n = m
		p = p[:m]
		err = ErrNoSpace
	}
	k := copy(b.data[b.front:], p)
	if k < n {
		copy(b.data, p[k:])
	}
	b.front = b.addIndex(b.front, n)
	return n, err
}

// WriteByte writes a single byte into the buffer. The error ErrNoSpace
// is returned if no single byte is available in the buffer for writing.
func (b *Buffer) WriteByte(c byte) error {
	if b.Available() < 1 {
		return ErrNoSpace
	}
	b.data[b.front] = c
	b.front = b.addIndex(b.front, 1)
	return nil
}

// prefixLen returns the length of the common prefix of a and b.
func prefixLen(a, b []byte) int {
	if len(a) > len(b) {
		a, b = b, a
	}
	for i, c := range a {
		if b[i] != c {
			return i
		}
	}
	return len(a)
}

// matchLen returns the length of the common prefix for the given
// distance from the rear and the byte slice p.
func (b *Buffer) MatchLen(distance int, p []byte) int {
	var n int
	i := b.rear - distance
	if i < 0 {
		if n = prefixLen(p, b.data[len(b.data)+i:]); n < -i {
			return n
		}
		p = p[n:]
		i = 0
	}
	n += prefixLen(p, b.data[i:])
	return n
}

func (b *Buffer) byteAt(i int) byte {
	if i < 0 {
		i += len(b.data)
	} else if i >= len(b.data) {
		i -= len(b.data)
	}

	return b.data[i]
}
