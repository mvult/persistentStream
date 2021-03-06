package main

import (
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/mvult/persistentStream/sender"
)

type testingSuite struct {
	reverseProxy *exec.Cmd
}

var testMaster testingSuite

func init() {
	testMaster = testingSuite{}
}

func TestMain(t *testing.T) {

	go func() {
		startReverseProxy()
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

	go mockReceiver("/test")

	tests := []string{"first", "second", "third"}

	var wg sync.WaitGroup
	sender.SetVerbose(true)
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
	stopReverseProxy()
}

func mockSender(testingID string) (*sender.PersistentStreamSender, error) {
	m := make(map[string]string)
	m["Persistent-Testing-ID"] = testingID

	u, _ := url.Parse("http://localhost:8082/test")

	return sender.SendStreamWithReader(getSource(), u, STREAM_BOUNDARY, m, nil, standardFunc)
}

func startReverseProxy() {
	switch runtime.GOOS {
	case "windows":
		testMaster.reverseProxy = exec.Command("./blankStream/reverseProxyServer/reverseProxyServer.exe")
	case "linux":
		testMaster.reverseProxy = exec.Command("./blankStream/reverseProxyServer/reverseProxyServer")
	case "darwin":
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
	if err := testMaster.reverseProxy.Process.Kill(); err != nil {
		panic(err)
	}
}
