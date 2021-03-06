package receiver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"sync"
	"time"

	"github.com/mvult/persistentStream/sender/globals"
)

const ITERATIONS_WAITING_LIMIT = 120

var logger *log.Logger
var verboseLog *log.Logger
var verbose bool

func init() {
	logger = log.New(os.Stdout, "", log.Llongfile|log.Ldate|log.Ltime)
	verbose = false
}

type PersistentStreamReceiver struct {
	id                    string
	writer                io.WriteCloser
	reader                io.ReadCloser
	inboundCompleteChan   chan bool
	waitingForReplacement bool
	responseWriter        http.ResponseWriter
	lock                  sync.Mutex
	boundary              string
}

func newReceiver(w http.ResponseWriter, r *http.Request, writerFunc func(w http.ResponseWriter, r *http.Request, mimeHeader textproto.MIMEHeader) (io.WriteCloser, error), boundary string) (*PersistentStreamReceiver, error) {
	id := r.Header.Get(globals.ID_HEADER)
	if id == "" {
		return &PersistentStreamReceiver{}, errors.New(fmt.Sprintf("Persistence stream doesn't include %v header with valid value", globals.ID_HEADER))
	}

	rdr, err := getMIMEMultipartReader(r, boundary)
	if err != nil {
		return &PersistentStreamReceiver{}, err
	}

	wrt, err := writerFunc(w, r, rdr.Header)
	if err != nil {
		return &PersistentStreamReceiver{}, err
	}
	ret := PersistentStreamReceiver{
		id:                    id,
		writer:                wrt,
		reader:                rdr,
		inboundCompleteChan:   make(chan bool),
		waitingForReplacement: false,
		responseWriter:        w,
		boundary:              boundary,
	}

	master.set(&ret)

	return &ret, nil
}

func HandlePersistentStream(w http.ResponseWriter, r *http.Request, boundary string, acceptFunc func(w http.ResponseWriter, r *http.Request) bool, writerFunc func(w http.ResponseWriter, r *http.Request, mimeHeader textproto.MIMEHeader) (io.WriteCloser, error)) {

	isCoord := r.Header.Get(globals.COORD_OR_STREAM_HEADER) == globals.COORD
	var b []byte

	if isCoord {
		if err := handleCoordination(w, r, acceptFunc); err != nil {
			b, _ = json.Marshal(globals.JsonResponse{Status: "failure", Error: err.Error()})
		} else {
			b, _ = json.Marshal(globals.JsonResponse{Status: "success"})
		}
	} else {
		if err := handleStream(w, r, writerFunc, boundary); err != nil {
			b, _ = json.Marshal(globals.JsonResponse{Status: "failure", Error: err.Error()})
		} else {
			b, _ = json.Marshal(globals.JsonResponse{Status: "success"})
		}
	}
	w.Write(b)
	logger.Printf("Responding to persistence request %v.  Is Coord: %v\n", string(b), isCoord)
	return
}

func handleCoordination(w http.ResponseWriter, r *http.Request, acceptFunc func(w http.ResponseWriter, r *http.Request) bool) error {

	isInitial := r.Header.Get(globals.INITIAL_OR_REATTACH_HEADER) == globals.INITIAL

	if isInitial {
		if acceptFunc(w, r) {
			return nil
		} else {
			return errors.New("Receiver not accepting streams")
		}
	} else {
		_, ok := master.get(r.Header.Get(globals.ID_HEADER))
		accepting := acceptFunc(w, r)

		if accepting && ok {
			return nil
		} else if !ok {
			return errors.New("Unable to accept reattach stream.  Stream ID not found.")
		} else {
			return errors.New("Receiver not accepting streams.")
		}
	}
}

func handleStream(w http.ResponseWriter, r *http.Request, writerFunc func(w http.ResponseWriter, r *http.Request, mimeHeader textproto.MIMEHeader) (io.WriteCloser, error), boundary string) (err error) {

	var pss *PersistentStreamReceiver
	defer func() {
		if err != nil {
			logger.Println("Fatal persistence error.  Closing downstream writer. Underlying error.", err)
			logger.Printf("PSS: %+v   Writer: %+v \n", pss, pss.writer)

			if pss.writer != nil {
				pss.writer.Close()
			}

			if pss.inboundCompleteChan != nil {
				close(pss.inboundCompleteChan)
			}
			master.delete(pss.id)
		}
	}()

	isInitial := r.Header.Get(globals.INITIAL_OR_REATTACH_HEADER) == globals.INITIAL

	if !isInitial {
		err = reattach(w, r)
		return
	} else {
		pss, err = newReceiver(w, r, writerFunc, boundary)
		if err != nil {
			return
		}

		iterationsWaiting := 0

		for {
			if pss.getWaitingForReplacement() {
				if iterationsWaiting > ITERATIONS_WAITING_LIMIT {
					err = fmt.Errorf("Waited too long for replacement stream %v\n", pss.id)
					return
				}

				iterationsWaiting++
				if iterationsWaiting < 5 || iterationsWaiting%10 == 0 {
					logger.Printf("Waiting for replacement on stream %v. Iterations Waiting: %v\n", pss.id, iterationsWaiting)
				}
				time.Sleep(time.Second * 5)
				continue
			}
			iterationsWaiting = 0

			if err = pss.writeToDownstream(); err != nil {
				if maybeNetworkError(err) {
					pss.setWaitingForReplacement(true)
					continue
				} else {
					return
				}
			} else {
				master.delete(pss.id)
				return
			}
		}
	}
}

func getMIMEMultipartReader(r *http.Request, boundary string) (*multipart.Part, error) {
	partRdr := multipart.NewReader(r.Body, boundary)
	return partRdr.NextPart()
}

func SetVerbose(v bool) {
	verbose = v
	if v {
		f, err := os.Create("persistentStreamReceiver.log")
		if err != nil {
			logger.Println(err)
			return
		}
		verboseLog = log.New(f, "", log.Ldate|log.Ltime)
	}
}
