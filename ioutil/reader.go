package ioutil

import (
	"io"
	"sync/atomic"
)

// ReaderFunc is a function type that implements the io.Reader interface.
type ReaderFunc func([]byte) (int, error)

func (f ReaderFunc) Read(p []byte) (int, error) { return f(p) }

// RoundRobinReader returns an io.Reader which round-robins across the given
// io.Readers when read.
// TODO: Fix EOFs
func RoundRobinReader(rs ...io.Reader) io.Reader {
	var robin uint64
	return ReaderFunc(func(p []byte) (n int, err error) {
		return rs[atomic.AddUint64(&robin, 1)%uint64(len(rs))].Read(p)
	})
}
