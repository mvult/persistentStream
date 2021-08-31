package main

import (
	"bytes"
	"fmt"
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
	fmt.Println(len(b))
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

// func getSource() io.ReadCloser {
// 	f, err := os.Create("./blankStream/log.log")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer f.Close()

// 	name := os.Args[1]
// 	fmt.Fprint(f, name)
// 	source, err := os.Open("./blankStream/SampleSession.h264")
// 	if err != nil {
// 		fmt.Fprint(f, err)
// 		panic(err)
// 	}

// 	dest := mockDest{}

// 	b := make([]byte, 1024*40)

// 	go func() {
// 		defer source.Close()
// 		var n int
// 		for {
// 			n, err = source.Read(b)
// 			if err != nil {
// 				if err == io.EOF {
// 					if _, err = dest.Write(b[:n]); err != nil {
// 						fmt.Fprint(f, err)
// 						panic(err)
// 					}
// 					break
// 				}
// 				fmt.Fprint(f, err)
// 				panic(err)
// 			}

// 			if _, err = dest.Write(b[:n]); err != nil {
// 				fmt.Fprint(f, err)
// 				panic(err)
// 			}
// 			time.Sleep(time.Second * 1)
// 		}
// 		fmt.Println("Sample Session Video complete")
// 	}()

// 	return dest
// }
