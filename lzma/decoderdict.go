// Copyright 2014-2019 Ulrich Kunitz. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lzma

import (
	"errors"

	"github.com/ulikunitz/xz/internal/buffer"
)

// decoderDict provides the dictionary for the decoder. The whole
// dictionary is used as reader buffer.
type decoderDict struct {
	buf  buffer.DecBuf
	head int64
}

// newDecoderDict creates a new decoder dictionary. The whole dictionary
// will be used as reader buffer.
func newDecoderDict(dictCap int) (d *decoderDict, err error) {
	// lower limit supports easy test cases
	if !(1 <= dictCap && int64(dictCap) <= MaxDictCap) {
		return nil, errors.New("lzma: dictCap out of range")
	}
	d = &decoderDict{buf: buffer.NewBuffer(dictCap)}
	return d, nil
}

// Reset clears the dictionary. The read buffer is not changed, so the
// buffered data can still be read.
func (d *decoderDict) Reset() {
	d.head = 0
}

// WriteByte writes a single byte into the dictionary. It is used to
// write literals into the dictionary.
func (d *decoderDict) WriteByte(c byte) error {
	if err := d.buf.WriteByte(c); err != nil {
		return err
	}
	d.head++
	return nil
}

// pos returns the position of the dictionary head.
func (d *decoderDict) pos() int64 { return d.head }

// dictLen returns the actual length of the dictionary.
func (d *decoderDict) dictLen() int {
	capacity := d.buf.Cap()
	if d.head >= int64(capacity) {
		return capacity
	}
	return int(d.head)
}

// byteAt returns a byte stored in the dictionary. If the distance is
// non-positive or exceeds the current length of the dictionary the zero
// byte is returned.
func (d *decoderDict) byteAt(dist int) byte {
	if !(0 < dist && dist <= d.dictLen()) {
		return 0
	}
	return buffer.DecByteAt(d.buf, dist)
}

// writeMatch writes the match at the top of the dictionary. The given
// distance must point in the current dictionary and the length must not
// exceed the maximum length 273 supported in LZMA.
//
// The error value ErrNoSpace indicates that no space is available in
// the dictionary for writing. You need to read from the dictionary
// first.
func (d *decoderDict) writeMatch(dist int64, length int) error {
	if !(0 < dist && dist <= int64(d.dictLen())) {
		return errors.New("writeMatch: distance out of range")
	}
	if !(0 < length && length <= maxMatchLen) {
		return errors.New("writeMatch: length out of range")
	}
	if length > d.buf.Available() {
		return buffer.ErrNoSpace
	}
	d.head += int64(length)

	return buffer.WriteMatch(d.buf, dist, length)
}

// Write writes the given bytes into the dictionary and advances the
// head.
func (d *decoderDict) Write(p []byte) (n int, err error) {
	n, err = d.buf.Write(p)
	d.head += int64(n)
	return n, err
}

// Available returns the number of available bytes for writing into the
// decoder dictionary.
func (d *decoderDict) Available() int { return d.buf.Available() }

// Read reads data from the buffer contained in the decoder dictionary.
func (d *decoderDict) Read(p []byte) (n int, err error) { return d.buf.Read(p) }

// Buffered returns the number of bytes currently buffered in the
// decoder dictionary.
func (d *decoderDict) buffered() int { return d.buf.Buffered() }

// Peek gets data from the buffer without advancing the rear index.
func (d *decoderDict) peek(p []byte) (n int, err error) { return d.buf.Peek(p) }
