package main

import (
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
)

const (
	crlf       = "\r\n"
	colonspace = ": "
)

func ChecksumMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sortedKeys := make([]string, 0, len(r.Header))
		canonRes := ""

		//use httptest's ResponseRecorder to extract status code, body and headers from an identical request

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, r)

		statusCode := strconv.Itoa(rr.Code)
		body := rr.Body.String()

		//lexographically sort header values

		for key := range rr.Header() {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sortedKeys)

		//construct the canonical response string

		canonRes = statusCode + crlf
		for _, header := range sortedKeys {
			canonRes += header + colonspace + rr.Header().Get(header) + crlf
		}

		canonRes += "X-Checksum-Headers" + colonspace
		canonRes += strings.Join(sortedKeys, ";")
		canonRes += crlf + crlf
		canonRes += body

		//calculate sha1 value of the canonical response, hex encode the result and add it as a header

		hsh := sha1.New()
		hsh.Write([]byte(canonRes))

		encodedCanonRes := hex.EncodeToString(hsh.Sum(nil))

		fmt.Println(encodedCanonRes)

		w.Header().Set("X-Checksum", encodedCanonRes)

		h.ServeHTTP(w, r)
	})
}

// Do not change this function.
func main() {
	var listenAddr = flag.String("http", ":8080", "address to listen on for HTTP")
	flag.Parse()

	http.Handle("/", ChecksumMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Foo", "bar")
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Date", "Sun, 08 May 2016 14:04:53 GMT")
		msg := "Curiosity is insubordination in its purest form.\n"
		w.Header().Set("Content-Length", strconv.Itoa(len(msg)))
		fmt.Fprintf(w, msg)
	})))

	log.Println(http.ListenAndServe(*listenAddr, nil))
}
