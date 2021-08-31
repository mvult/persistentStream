package main

import (
	"io"
	"net/http"
	"os"
	"persistentStream/receiver"
)

func mockReceiver() {
	http.HandleFunc("/test", handlePersistence)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func handlePersistence(w http.ResponseWriter, r *http.Request) {
	receiver.HandlePersistentStream(w, r, STREAM_BOUNDARY, acceptFunc, writerFunc)
}

func acceptFunc(w http.ResponseWriter, r *http.Request) bool {
	return accepting
}

func writerFunc(w http.ResponseWriter, r *http.Request) (io.WriteCloser, error) {
	f, err := os.Create(r.Header.Get("Persistent-Testing-ID"))
	return f, err
}
func standardFunc(res *http.Response, err error) {

}
