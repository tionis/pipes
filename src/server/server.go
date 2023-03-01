package server

import (
	"bytes"
	_ "embed"
	"io"
	"net/http"
	"sync"
	log "tasadar.net/tionis/pipes/src/logger"
)

//go:embed index.html
var indexBytes []byte

//go:embed app.css
var cssBytes []byte

// stream contains the reader and the channel to signify its read
type stream struct {
	reader io.ReadCloser
	done   chan struct{}
}

// Serve will start the server
func Serve(flagPort string) (err error) {
	channels := make(map[string]chan stream)
	mutex := &sync.Mutex{}

	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("opened %s %s", r.Method, r.URL.Path)
		defer func() {
			log.Debugf("finished %s\n", r.URL.Path)
		}()

		switch r.URL.Path {
		case "/":
			// serve the index page
			w.Header().Add("Content-Type", "text/html")
			_, err := w.Write(indexBytes)
			if err != nil {
				log.Error(err)
				return
			}
			return
		case "/app.css":
			// serve the index page
			w.Header().Add("Content-Type", "text/css")
			_, err := w.Write(cssBytes)
			if err != nil {
				log.Error(err)
				return
			}
			return
		}

		mutex.Lock()
		if _, ok := channels[r.URL.Path]; !ok {
			channels[r.URL.Path] = make(chan stream)
		}
		channel := channels[r.URL.Path]
		mutex.Unlock()

		queries, ok := r.URL.Query()["pubsub"]
		pubsub := ok && queries[0] == "true"
		log.Debugf("pubsub: %+v", pubsub) // TODO switch to path based behaviour switching
		// TODO add following path prefixes for different behaviours:
		// /pubsub/** - non blocking, readers will only receive this when they are listening at the moment
		// /fifo/** - blocking pubsub
		// /req/** - requests
		// /res/** - responses
		// /steams/ - streams (similar to redis streams) (probably out-of-scope, was just an idea)

		method := r.Method
		queries, ok = r.URL.Query()["body"]
		var bodyString string
		if ok {
			bodyString = queries[0]
			if bodyString != "" {
				method = "POST"
			}
		}

		log.Debug(channel)
		if method == "GET" {
			select {
			case stream := <-channel:
				_, err := io.Copy(w, stream.reader)
				if err != nil {
					return
				}
				close(stream.done)
			case <-r.Context().Done():
				log.Debug("consumer canceled")
			}
		} else if method == "POST" {
			var buf []byte
			if bodyString != "" {
				buf = []byte(bodyString)
			} else {
				buf, _ = io.ReadAll(r.Body)
			}

			if !pubsub {
				log.Debug("no pubsub POST")
				doneSignal := make(chan struct{})
				stream := stream{reader: io.NopCloser(bytes.NewBuffer(buf)), done: doneSignal}
				select {
				case channel <- stream:
					log.Debug("connected to consumer")
				case <-r.Context().Done():
					log.Debug("producer canceled")
					doneSignal <- struct{}{}
				}
				log.Debug("waiting for done")
				<-doneSignal
			} else {
				defer func() {
					log.Debug("finished pubsub")
				}()
				log.Debug("using pubsub")
				finished := false
				for {
					if finished {
						break
					}
					doneSignal := make(chan struct{})
					stream := stream{reader: io.NopCloser(bytes.NewBuffer(buf)), done: doneSignal}
					select {
					case channel <- stream:
						log.Debug("connected to consumer")
					case <-r.Context().Done():
						log.Debug("producer canceled")
					default:
						log.Debug("no one connected")
						close(doneSignal)
						finished = true
					}
					<-doneSignal
				}
			}
		}
	}

	log.Infof("running on port %s", flagPort)
	err = http.ListenAndServe(":"+flagPort, http.HandlerFunc(handler))
	if err != nil {
		log.Error(err)
	}
	return
}
