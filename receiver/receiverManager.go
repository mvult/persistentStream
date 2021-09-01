package receiver

import (
	"errors"
	"sync"
)

var master receiverManager

func init() {
	master = receiverManager{streams: make(map[string]*PersistentStreamReceiver)}
}

type receiverManager struct {
	lock    sync.Mutex
	streams map[string]*PersistentStreamReceiver
}

func (rm *receiverManager) get(s string) (*PersistentStreamReceiver, bool) {
	rm.lock.Lock()
	defer rm.lock.Unlock()
	ret, ok := rm.streams[s]
	return ret, ok
}

func (rm *receiverManager) set(psr *PersistentStreamReceiver) {
	rm.lock.Lock()
	defer rm.lock.Unlock()
	_, ok := rm.streams[psr.id]
	if ok {
		panic(errors.New("Should be impossible to have the same id"))
	}

	rm.streams[psr.id] = psr
}

func (rm *receiverManager) delete(s string) {
	rm.lock.Lock()
	defer rm.lock.Unlock()
	logger.Println("Deleting", s)
	delete(rm.streams, s)
}
