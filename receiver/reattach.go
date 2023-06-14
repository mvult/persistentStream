package receiver

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mvult/persistentStream/sender/globals"
)

func reattach(w http.ResponseWriter, r *http.Request) error {
	id := r.Header.Get(globals.ID_HEADER)
	if id == "" {
		return errors.New(fmt.Sprintf("Invalid %v header in reattach request", globals.ID_HEADER))
	}

	pss, ok := master.get(id)
	if !ok {
		logger.Printf("Unable to find stream id %v.  Current master: %+v\n", id, master.streams)
		return fmt.Errorf("Unable to find stream for id %v\n", id)
	}
	if err := pss.replaceRequest(w, r); err != nil {
		return err
	}
	pss.setWaitingForReplacement(false)

	logger.Printf("Successfully reattached stream %v.  Current master: %v\n", id, master.streams)

	<-pss.inboundCompleteChan
	return nil
}

func (pss *PersistentStreamReceiver) setWaitingForReplacement(val bool) {
	pss.lock.Lock()
	defer pss.lock.Unlock()
	pss.waitingForReplacement = val
}

func (pss *PersistentStreamReceiver) getWaitingForReplacement() bool {
	pss.lock.Lock()
	defer pss.lock.Unlock()
	return pss.waitingForReplacement
}

func (pss *PersistentStreamReceiver) replaceRequest(w http.ResponseWriter, r *http.Request) error {
	pss.lock.Lock()
	defer pss.lock.Unlock()

	part, err := getMIMEMultipartReader(r, pss.boundary)
	if err != nil {
		return err
	}
	pss.reader = part
	pss.responseWriter = w
	return nil
}
