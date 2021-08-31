package timeoutWriter

import (
	"errors"
	"io"
	"log"
	"os"
	"time"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "", log.Llongfile|log.Ldate|log.Ltime)
}

var WRITER_TIMEOUT_ERROR = errors.New("WRITER_TIMEOUT_ERROR")

type timeoutWriterCloser struct {
	writerCloser io.WriteCloser
	timeout      time.Duration
}

func NewTimeoutWriterCloser(wc io.WriteCloser, timeout time.Duration) timeoutWriterCloser {
	return timeoutWriterCloser{writerCloser: wc, timeout: timeout}
}

func (tw timeoutWriterCloser) Close() error {
	return tw.writerCloser.Close()
}

func (tw timeoutWriterCloser) Write(p []byte) (n int, err error) {

	defer func() {
		if r := recover(); r != nil {
			err, _ = r.(error)
			logger.Println(err)
		}
	}()

	writeChan := make(chan int)

	go func() {
		n, err = tw.writerCloser.Write(p)

		if err != nil {
			logger.Println(err)
		}

		writeChan <- n
	}()

	select {
	case <-time.NewTicker(tw.timeout).C:
		logger.Println("Timing out stream write")
		n = 0
		err = WRITER_TIMEOUT_ERROR
		return
	case <-writeChan:
		return
	}
}
