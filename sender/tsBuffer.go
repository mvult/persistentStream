package sender

import (
	"bytes"
	"sync"
)

type Buffer struct {
	B      bytes.Buffer
	M      sync.Mutex
	Closed bool
}

func (b *Buffer) Close() error {
	b.M.Lock()
	defer b.M.Unlock()
	b.Closed = true
	return nil
}

func (b *Buffer) IsClosed() bool {
	b.M.Lock()
	defer b.M.Unlock()
	return b.Closed
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	b.M.Lock()
	defer b.M.Unlock()
	return b.B.Read(p)
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	b.M.Lock()
	defer b.M.Unlock()
	return b.B.Write(p)
}

func (b *Buffer) Len() (n int) {
	b.M.Lock()
	defer b.M.Unlock()
	return b.B.Len()
}
