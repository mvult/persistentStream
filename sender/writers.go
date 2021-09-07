package sender

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mvult/persistentStream/sender/globals"
	"github.com/mvult/persistentStream/sender/timeoutWriter"

	"github.com/mvult/httpStreamWriter"
)

const MAX_RECONNECT_ATTEMPTS = 30
const MAX_GET_WRITER_ATTEMPS = 90

func (pss *PersistentStreamSender) SendBufferToHttp() (err error) {
	// send initial coords message
	if err = pss.receiverAccepting(false); err != nil {
		return
	}

	if err = pss.getHttpWriter(false); err != nil {
		return
	}

	logger.Println("Sending to HTTP")

	go func() {
		numReconnectAttemps := 0
		var remnantBytes []byte

		for {
			if numReconnectAttemps > 0 {
				if err = pss.getReattachedWriter(); err != nil {
					logger.Println(err)
					pss.TerminateOnError(err)
					return
				}
			}
			numReconnectAttemps++
			remnantBytes, err = pss.writeToHttp(remnantBytes)
			if err == nil {
				logger.Println("Succesfully completed HTTP send")
				pss.completeChan <- true
				return
			}

			if !isNetworkError(err) {
				err = fmt.Errorf("Unknown HTTP error: %v.  Terminating persistent send. \n", err.Error())
				logger.Println(err)
				pss.TerminateOnError(err)
				return
			}

			if numReconnectAttemps > MAX_RECONNECT_ATTEMPTS {
				err = errors.New("Persistent stream reached reconnect attempt limit")
				logger.Println(err)
				pss.TerminateOnError(err)
				return
			}

			logger.Printf("Possible network error: %v.  ATTEMPTING TO RECONNECT IN 5 seconds.  Attempt %v of %v\n", err.Error(), numReconnectAttemps, MAX_RECONNECT_ATTEMPTS)
			time.Sleep(time.Second * 5)
		}
	}()
	return nil
}

func (pss *PersistentStreamSender) getHttpWriter(reattach bool) error {
	tmpHeaders := make(map[string]string)
	// tmpHeaders := pss.connectionParams.httpHeaders
	for k, h := range pss.connectionParams.httpHeaders {
		tmpHeaders[k] = h
	}

	tmpHeaders[globals.COORD_OR_STREAM_HEADER] = globals.STREAM
	tmpHeaders[globals.ID_HEADER] = pss.id

	if reattach {
		tmpHeaders[globals.INITIAL_OR_REATTACH_HEADER] = globals.REATTACH
	} else {
		tmpHeaders[globals.INITIAL_OR_REATTACH_HEADER] = globals.INITIAL
	}

	wrt, err := httpStreamWriter.HttpStreamWriter(pss.connectionParams.target,
		pss.connectionParams.boundary,
		tmpHeaders,
		pss.connectionParams.mimeHeaders,
		pss.connectionParams.callback,
	)

	if err != nil {
		return err
	}
	pss.writer = timeoutWriter.NewTimeoutWriterCloser(wrt, time.Second*10)
	return nil

}

func (pss *PersistentStreamSender) getReattachedWriter() error {
	numGetWriterAttempts := 0
	var err error

	for {
		if err = pss.receiverAccepting(true); err != nil {
			goto errorHandle
		}

		err = pss.getHttpWriter(true)

	errorHandle:
		if err == nil {
			logger.Println("Successfully reattached")
			return err
		} else {
			logger.Printf("Error getting http writer. Underlying error: '%v'.  Attemps: %v out of %v\n", err, numGetWriterAttempts, MAX_GET_WRITER_ATTEMPS)
			numGetWriterAttempts++

			if numGetWriterAttempts > MAX_GET_WRITER_ATTEMPS {
				return errors.New("Persistent stream reached get writer attempt limit")
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func isNetworkError(err error) bool {
	// return strings.Contains(err.Error(), "io: read/write on closed pipe") || strings.Contains(err.Error(), WRITER_TIMEOUT_ERROR.Error())
	return strings.Contains(err.Error(), "io: read/write on closed pipe") || errors.Is(err, timeoutWriter.WRITER_TIMEOUT_ERROR)
}
