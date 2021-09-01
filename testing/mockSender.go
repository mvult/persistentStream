package main

import (
	"bytes"
	"io"
	"os"
	"time"
)

type mockDest struct {
	b bytes.Buffer
}

func (m mockDest) Close() error {
	return nil
}

func (m mockDest) Read(b []byte) (int, error) {
	return m.b.Read(b)
}

func (m mockDest) Write(b []byte) (int, error) {
	return m.b.Write(b)
}

type slowFile struct {
	f *os.File
}

func (sf slowFile) Read(b []byte) (int, error) {
	time.Sleep(time.Second * 1)
	return sf.f.Read((b))
}

func (sf slowFile) Close() error {
	return sf.f.Close()
}

func getSource() io.ReadCloser {
	source, err := os.Open("./blankStream/SampleSession.h264")
	if err != nil {
		panic(err)
	}
	return slowFile{source}
}
