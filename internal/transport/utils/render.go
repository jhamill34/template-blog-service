package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func RenderJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_, _ = w.Write(buf.Bytes())
}

func RenderHTML(w http.ResponseWriter, doc string, statusCode int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(doc))
}
