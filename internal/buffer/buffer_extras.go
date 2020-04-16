package buffer

import (
	"fmt"
	"io"
)

func (buf *Buffer) ByteAt(i int) byte {
	if i < 0 {
		i += len(buf.data)
	} else if i >= len(buf.data) {
		i -= len(buf.data)
	}

	return buf.data[i]
}

func (buf *Buffer) ByteAtRP(rearPlus int) byte {
	return buf.ByteAt(buf.rear + rearPlus)
}

func (b *Buffer) DecByteAt(dist int) byte {
	return b.ByteAt(b.front - dist)
}

func (b *Buffer) EncByteAt(distance int) byte {
	return b.ByteAt(b.rear - distance)
}

func (b *Buffer) CopyN(w io.Writer, n int) (written int, err error) {
	i := b.rear - n
	var e error
	if i < 0 {
		i += len(b.data)
		if written, e = w.Write(b.data[i:]); e != nil {
			return written, e
		}
		i = 0
	}
	var k int
	k, e = w.Write(b.data[i:b.rear])
	written += k
	if e != nil {
		err = e
	}
	return written, err
}

func (b *Buffer) WriteMatch(dist int64, length int) error {
	i := b.front - int(dist)
	if i < 0 {
		i += len(b.data)
	}
	for length > 0 {
		var p []byte
		if i >= b.front {
			p = b.data[i:]
			i = 0
		} else {
			p = b.data[i:b.front]
			i = b.front
		}
		if len(p) > length {
			p = p[:length]
		}
		if _, err := b.Write(p); err != nil {
			panic(fmt.Errorf("b.Write returned error %s", err))
		}
		length -= len(p)
	}
	return nil
}
