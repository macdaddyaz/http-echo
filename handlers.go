package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"macdaddyaz/http-echo/version"
)

const (
	httpHeaderAppName    string = "X-App-Name"
	httpHeaderAppVersion string = "X-App-Version"

	httpLogDateFormat string = "2006/01/02 15:04:05"
	httpLogFormat     string = "%v %s %s \"%s %s %s\" %d %d \"%s\" %v\n"
)

// withAppHeaders adds application headers such as X-App-Version and X-App-Name.
func withAppHeaders(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(httpHeaderAppName, version.Name)
		w.Header().Set(httpHeaderAppVersion, version.Version)
		h(w, r)
	}
}

// metaResponseWriter is a response writer that saves information about the
// response for logging.
type metaResponseWriter struct {
	writer http.ResponseWriter
	status int
	length int
}

// Header implements the http.ResponseWriter interface.
func (w *metaResponseWriter) Header() http.Header {
	return w.writer.Header()
}

// WriteHeader implements the http.ResponseWriter interface.
func (w *metaResponseWriter) WriteHeader(s int) {
	w.status = s
	w.writer.WriteHeader(s)
}

// Write implements the http.ResponseWriter interface.
func (w *metaResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.length = len(b)
	return w.writer.Write(b)
}

// httpLog accepts an io object and logs the request and response objects to the
// given io.Writer.
func httpLog(out io.Writer, h http.HandlerFunc, stream bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var mrw metaResponseWriter
		mrw.writer = w

		defer func(start time.Time) {
			status := mrw.status
			length := mrw.length
			end := time.Now()
			dur := end.Sub(start)
			fmt.Fprintf(out, httpLogFormat,
				end.Format(httpLogDateFormat),
				r.Host, r.RemoteAddr, r.Method, r.URL.Path, r.Proto,
				status, length, r.UserAgent(), dur)
			if r.Body != nil {
				if stream {
					logBody(out, r.Body)
				} else {
					logBodyMem(out, r.Body)
				}
			}
		}(time.Now())

		h(&mrw, r)
	}
}

func logBody(out io.Writer, body io.Reader) {
	buf := bufio.NewReader(body)
	count, err := buf.WriteTo(out)
	if err != nil {
		log.Fatal(err)
	}
	if count > 0 {
		fmt.Fprintln(out)
	}
}

func logBodyMem(out io.Writer, body io.Reader) {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		log.Fatal(err)
	}
	buf := bufio.NewWriter(out)
	count, err := buf.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	if count > 0 {
		fmt.Fprintln(out)
	}
}
