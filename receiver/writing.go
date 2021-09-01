package receiver

import (
	"errors"
	"fmt"
	"io"
	"time"
)

const streamReaderTimeout = 45

var ConnectionResetByPeerError = errors.New("connection reset by peer")
var ReaderTimeOutError = errors.New("Reader Timeout")

func (pss *PersistentStreamReceiver) writeToDownstream() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err, _ = r.(error)
			logger.Println(err)
		}
	}()

	buf := make([]byte, 1024*1024)
	var n int

	for {
		n, err = pss.Read(buf)
		if err != nil {
			if err == io.EOF {
				logger.Println("Got to stream EOF")
				break
			}
			logger.Println(err)
			return err
		}

		if verbose {
			fmt.Println("Bytes read from sender", n)
		}

		if n, err = pss.Write(buf[:n]); err != nil {
			logger.Println(err)
			return err
		}

		if verbose {
			fmt.Println("Bytes written to destination", n)
		}
	}

	logger.Println("Closing writer", pss.id)
	if err = pss.writer.Close(); err != nil {
		return err
	}

	logger.Println("Closing Reader", pss.id)
	if err = pss.reader.Close(); err != nil {
		return err
	}

	logger.Println("Closing Inbound chan", pss.id)
	close(pss.inboundCompleteChan)
	return nil
}

func (pss *PersistentStreamReceiver) Read(p []byte) (n int, err error) {
	pss.lock.Lock()
	defer pss.lock.Unlock()

	defer func() {
		if r := recover(); r != nil {
			err, _ = r.(error)
			logger.Println(err)
		}
	}()

	readChan := make(chan int)

	go func() {
		n, err = pss.reader.Read(p)

		if err != nil {
			logger.Println(err)
		}

		readChan <- n
	}()

	select {
	case <-time.NewTicker(time.Second * streamReaderTimeout).C:
		logger.Println("Timing out stream read")
		n = 0
		err = ReaderTimeOutError
		return
	case <-readChan:
		return
	}
}

func (pss *PersistentStreamReceiver) Write(p []byte) (int, error) {
	return pss.writer.Write(p)
}

func maybeNetworkError(err error) bool {
	return true
}
