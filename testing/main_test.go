package main

import (
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"persistentStream/sender"
	"runtime"
	"sync"
	"testing"
	"time"
)

type testingSuite struct {
	reverseProxy *exec.Cmd
}

var testMaster testingSuite

func init() {
	testMaster = testingSuite{}
}

const STREAM_BOUNDARY = "JwnftdsGXBsijUljzOQsjqJmqZMvbGHqgxXn"

var accepting = true

func TestMain(t *testing.T) {
	go startReverseProxy()

	go func() {
		time.Sleep(10 * time.Second)
		fmt.Println("Stopping proxy")
		stopReverseProxy()
		time.Sleep(30 * time.Second)
		fmt.Println("Starting proxy")
		startReverseProxy()
		time.Sleep(10 * time.Second)
		fmt.Println("Stopping proxy")
		stopReverseProxy()
		time.Sleep(10 * time.Second)
		fmt.Println("Starting proxy")
		startReverseProxy()
	}()

	go mockReceiver()

	tests := []string{"first", "second", "third"}

	var wg sync.WaitGroup

	for _, s := range tests {
		time.Sleep(time.Second * 1)

		go func(s string) {
			wg.Add(1)
			fmt.Println("Starting stream", s)
			pss, err := mockSender(s)
			if err != nil {
				t.Error("Failed to send")
			}
			if err := pss.Wait(); err != nil {
				t.Error(err)
			}
			wg.Done()
		}(s)
	}
	wg.Wait()

}

func mockSender(testingID string) (*sender.PersistentStreamSender, error) {
	m := make(map[string]string)
	m["Persistent-Testing-ID"] = testingID

	u, _ := url.Parse("http://localhost:8082/test")

	return sender.SendStream(getSource(), u, STREAM_BOUNDARY, m, nil, standardFunc)
}

func startReverseProxy() {
	switch runtime.GOOS {
	case "windows":
		testMaster.reverseProxy = exec.Command("./blankStream/reverseProxyServer/reverseProxyServer.exe")
	case "linux":
		testMaster.reverseProxy = exec.Command("./blankStream/reverseProxyServer/reverseProxyServer")
	default:
		log.Fatal("OS not supported")
	}

	err := testMaster.reverseProxy.Start()
	if err != nil {
		panic(err)
	}
}

func stopReverseProxy() {
	testMaster.reverseProxy.Process.Kill()
}
