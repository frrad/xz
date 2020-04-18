package buffer

import (
	"fmt"
	"io"
)

func (b *Buffer) EncByteAt(dist int) byte {
	return b.byteAt(b.rear + dist)
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

func WriteMatch(b DecBuf, dist int64, length int) error {
	for length > 0 {
		p, err := b.PeekTail(dist, length)
		if err != nil {
			return err
		}

		written, err := b.Write(p)
		if err != nil {
			panic(fmt.Errorf("b.Write returned error %s", err))
		}
		if written != len(p) {
			panic(fmt.Errorf("didn't write entire buffer:  %s", err))
		}

		length -= written
	}

	return nil
}

// DecBytAt retrieves the byte dist bytes from the end of buffered data
func DecByteAt(b DecBuf, dist int) byte {
	bs, err := b.PeekTail(int64(dist), 1)
	if err != nil {
		panic(err)
	}
	return bs[0]
}
