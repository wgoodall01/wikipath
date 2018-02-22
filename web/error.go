package main

import (
	"encoding/json"
	"net/http"
)

type HttpError struct {
	Status  int    `json:"status"` // Status code
	kind    string // Unique kind
	Message string `json:"message"` // Descriptive message
}

func NewHttpError(status int, message string) *HttpError {
	return &HttpError{
		Status:  status,
		Message: message,
	}
}

func (he *HttpError) Send(w http.ResponseWriter) {
	w.WriteHeader(he.Status)
	bytes, _ := json.MarshalIndent(he, "", "  ")
	w.Write(bytes)
}
