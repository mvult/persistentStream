package receiver

import (
	"fmt"
)

type PersistenceTimeoutError struct {
	id string
}

func (e *PersistenceTimeoutError) Error() string {
	return fmt.Sprintf("Persistance Timeout Error on stream id: %v", e.id)
}
