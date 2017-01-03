package handler

import (
	"bytes"
	"io"
	"net/http"
)

type BufferedResponseWriter struct {
	http.ResponseWriter
	buffer *bytes.Buffer
	status int
}

func (w *BufferedResponseWriter) Write(data []byte) (int, error) {
	return w.buffer.Write(data)
}

func (w *BufferedResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *BufferedResponseWriter) Send() {
	w.ResponseWriter.WriteHeader(w.status)
	io.Copy(w.ResponseWriter, w.buffer)
}

func NewBufferedResponseWriter(rw http.ResponseWriter) *BufferedResponseWriter {
	return &BufferedResponseWriter{
		rw,
		&bytes.Buffer{},
		http.StatusOK,
	}
}
