package sender

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mvult/persistentStream/sender/globals"
	// "persistentStream/sender/globals"
)

func (psw *PersistentStreamSender) receiverAccepting(reattach bool) (error, error) {
	req, err := http.NewRequest("POST", psw.connectionParams.target.String(), nil)
	if err != nil {
		logger.Println(err)
		return err, nil
	}

	req.Header = make(map[string][]string)

	req.Header[globals.ID_HEADER] = []string{psw.id}
	req.Header[globals.COORD_OR_STREAM_HEADER] = []string{globals.COORD}
	if reattach {
		req.Header[globals.INITIAL_OR_REATTACH_HEADER] = []string{globals.REATTACH}
	} else {
		req.Header[globals.INITIAL_OR_REATTACH_HEADER] = []string{globals.INITIAL}
	}

	for k, h := range psw.connectionParams.httpHeaders {
		req.Header[k] = []string{h}
	}

	client := &http.Client{Timeout: 0}

	res, err := client.Do(req)
	if err != nil {
		logger.Println(err)
		return err, nil
	}

	var respJson globals.JsonResponse
	if err := json.NewDecoder(res.Body).Decode(&respJson); err != nil {
		logger.Println(err)
		return nil, fmt.Errorf("Unable to parse coordination message.  Reattach: %v", reattach)
	}
	logger.Println(respJson)

	if respJson.Status != "success" {
		err = fmt.Errorf(respJson.Error)
		logger.Println(err)
		return nil, err
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("Unknown error.  Status: %v   Error: %v", res.StatusCode, respJson.Error)
		logger.Println(err)
		return nil, err
	}

	return nil, nil
}
