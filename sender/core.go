package sender

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/google/uuid"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "", log.Llongfile|log.Ldate|log.Ltime)
}

type PersistentStreamSender struct {
	writer           io.WriteCloser
	buffer           Buffer
	errorChan        chan error
	completeChan     chan bool
	id               string
	connectionParams connectionParams
}

type connectionParams struct {
	target      *url.URL
	boundary    string
	httpHeaders map[string]string
	mimeHeaders map[string]string
	callback    func(r *http.Response, err error)
}

func SendStreamWithReader(source io.ReadCloser, target *url.URL, boundary string, httpHeaders map[string]string, mimeHeaders map[string]string, responseFunc func(r *http.Response, err error)) (*PersistentStreamSender, error) {

	var err error

	ret := PersistentStreamSender{
		buffer:       Buffer{},
		id:           uuid.New().String(),
		errorChan:    make(chan error, 1),
		completeChan: make(chan bool, 1),
		connectionParams: connectionParams{
			target:      target,
			boundary:    boundary,
			httpHeaders: httpHeaders,
			mimeHeaders: mimeHeaders,
			callback:    responseFunc,
		}}

	err = ret.SendBufferToHttp()

	if err == nil {
		go ret.writeSourceToBuffer(source)
	}
	return &ret, err
}

func SendStreamWithWriter(target *url.URL, boundary string, httpHeaders map[string]string, mimeHeaders map[string]string, responseFunc func(r *http.Response, err error)) (*PersistentStreamSender, error) {

	var err error

	ret := PersistentStreamSender{
		buffer:       Buffer{},
		id:           uuid.New().String(),
		errorChan:    make(chan error, 1),
		completeChan: make(chan bool, 1),
		connectionParams: connectionParams{
			target:      target,
			boundary:    boundary,
			httpHeaders: httpHeaders,
			mimeHeaders: mimeHeaders,
			callback:    responseFunc,
		}}

	err = ret.SendBufferToHttp()

	return &ret, err
}

func (pss *PersistentStreamSender) Write(b []byte) (int, error) {
	return pss.buffer.Write(b)
}

func (pss *PersistentStreamSender) Close() error {
	return pss.buffer.Close()
}

func (psw *PersistentStreamSender) writeSourceToBuffer(source io.ReadCloser) {
	var n int
	var err error
	buf := make([]byte, 1024*1024)

	for {
		n, err = source.Read(buf)

		if err == io.EOF {
			if _, err = psw.buffer.Write(buf[:n]); err != nil {
				logger.Panicln(err)
			}
			psw.buffer.Close()

			if err := source.Close(); err != nil {
				logger.Println("Not fatal error:", err)
			}

			return
		}
		if err != nil {
			logger.Panicln(err)
		}

		if _, err = psw.buffer.Write(buf[:n]); err != nil {
			logger.Panicln(err)
		}
	}
}

func (pss *PersistentStreamSender) Wait() error {
	select {
	case err := <-pss.errorChan:
		return err
	case <-pss.completeChan:
		return nil
	}
}
