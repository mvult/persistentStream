package sender

import (
	"io"
	"time"
)

func (pss *PersistentStreamSender) writeToHttp(remnantBytes []byte) (newRemnantBytes []byte, err error) {
	var n int
	buf := make([]byte, 1024)

	defer func() {
		if r := recover(); r != nil {
			newRemnantBytes = buf[:n]
			err, _ = r.(error)
			logger.Printf("In recover func.  err: %v   newRemnantBytes: %v\n", err, newRemnantBytes)
		}
	}()

	if len(remnantBytes) > 0 {
		_, err := pss.writer.Write(remnantBytes)
		if err != nil {
			logger.Println(err)
			return buf[:n], err
		}
	}

	for {
		n, err = pss.buffer.Read(buf)
		if err == io.EOF {

			if pss.buffer.IsClosed() {
				if n != 0 {
					if _, err := pss.writer.Write(buf[:n]); err != nil {
						logger.Println(err)
						return buf[:n], err
					}
				}

				if err = pss.writer.Close(); err != nil {
					logger.Println("Error closing persistent writer", err)
				}
				logger.Println("Closed HTTP send body")

				return buf[:n], nil
			}
			time.Sleep(time.Millisecond * 100)
			continue
		}

		if err != nil {
			logger.Println(err)
			return buf[:n], err
		}

		_, err := pss.writer.Write(buf[:n])
		if err != nil {
			logger.Println(err)
			return buf[:n], err
		}
	}
}
