package persistentStream

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/mvult/persistentStream/sender"
)

func init() {
	u, _ := url.Parse("www.google.com")
	f, _ := os.Create("Hello")
	sender.SendStream(f, u, "923423", nil, nil, StandardResponse)
}

func StandardResponse(res *http.Response, err error) {
	if res == nil || err != nil {
		log.Printf("Error on stream send.  %v\n", err)
		return
	}

	res_content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	if err = res.Body.Close(); err != nil {
		panic(err)
	}
	if len(res_content) != 0 {
		log.Println(string(res_content))
	}
}
