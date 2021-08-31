package main

import (
	"net/url"
	"os/exec"
	"persistentStream/sender"
	"syscall"
	"testing"
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
	// go startReverseProxy()

	// go func() {
	// 	time.Sleep(5 * time.Second)
	// 	fmt.Println("Stopping proxy")
	// 	stopReverseProxy()
	// 	time.Sleep(20 * time.Second)
	// 	fmt.Println("Starting proxy")
	// 	startReverseProxy()
	// }()

	go mockReceiver()
	pss, err := mockSender()
	if err != nil {
		t.Error("Failed to send")
	}
	if err := pss.Wait(); err != nil {
		t.Error(err)
	}

}

func mockSender() (*sender.PersistentStreamSender, error) {
	m := make(map[string]string)
	m["Persistent-Testing-ID"] = "osjdifjofelkejlf"

	u, _ := url.Parse("http://localhost:8082/test")

	return sender.SendStream(getSource(), u, STREAM_BOUNDARY, m, nil, standardFunc)
}

func startReverseProxy() {
	testMaster.reverseProxy = exec.Command("./blankStream/reverseProxyServer/reverseProxyServer.exe")
	err := testMaster.reverseProxy.Start()
	if err != nil {
		panic(err)
	}
}

func stopReverseProxy() {
	testMaster.reverseProxy.Process.Signal(syscall.SIGINT)
}
