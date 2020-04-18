package buffer

import (
	"bytes"
	"fmt"
)

type OnDemand struct {
	size       int
	readOffset int
	buf        *bytes.Buffer
}

func NewOnDemandBuffer(n int) *OnDemand {
	return &OnDemand{
		size: n,
		buf:  &bytes.Buffer{},
	}
}

func (o *OnDemand) Cap() int {
	return o.size
}

func (o *OnDemand) Buffered() int {
	return o.buf.Len() - o.readOffset
}

func (o *OnDemand) Available() int {
	return o.Cap() - o.Buffered()
}

func (o *OnDemand) PeekTail(dist int64, length int) ([]byte, error) {
	slice := o.buf.Bytes()

	if int(dist) > len(slice) {
		dist -= int64(len(slice))

		return []byte{}, fmt.Errorf("can't read %d bytes into buf of len %d", dist, len(slice))
	}

	if int64(length) > dist {
		length = int(dist)
	}

	startOffset := len(slice) - int(dist)

	return slice[startOffset : startOffset+length], nil
}

func (o *OnDemand) Peek(p []byte) (int, error) {
	m := o.Buffered()
	n := len(p)
	if m < n {
		n = m
		p = p[:n]
	}

	copy(p, o.buf.Bytes()[o.readOffset:o.readOffset+n])
	return n, nil
}

func (o *OnDemand) Read(p []byte) (int, error) {
	n, err := o.Peek(p)
	o.readOffset += n
	return n, err
}

func (o *OnDemand) Write(p []byte) (int, error) {
	m := o.Available()
	n := len(p)
	var spaceErr error
	if m < n {
		n = m
		p = p[:m]
		spaceErr = ErrNoSpace
	}

	written, writeError := o.buf.Write(p)

	if writeError != nil {
		return written, writeError
	}

	o.freeOldBytes()
	return written, spaceErr
}

func (o *OnDemand) WriteByte(c byte) error {
	if o.Available() < 1 {
		return ErrNoSpace
	}

	o.freeOldBytes()
	return o.buf.WriteByte(c)
}

func (o *OnDemand) freeOldBytes() {
	excessBytes := o.buf.Len() - o.size

	if excessBytes <= 0 {
		return
	}

	o.buf.Next(excessBytes)
	o.readOffset -= excessBytes
}
