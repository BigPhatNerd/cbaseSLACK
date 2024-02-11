// middleware.go
package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the body
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		r.Body.Close() // ensure the original body is closed

		// Log the request details
		logger.WithFields(logrus.Fields{
			"method":  r.Method,
			"path":    r.URL.Path,
			"query":   r.URL.Query().Encode(),
			"body":    string(bodyBytes),
			"ip":      r.RemoteAddr,
		}).Info("HTTP request information")

		// Create a new ReadCloser for the original body
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
