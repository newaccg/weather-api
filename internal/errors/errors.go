package errors

import (
	"encoding/json"
	"log"
	"net/http"
)

type Error struct{
	InternalError error
	ErrorMessage string
	HttpCode int
}

func NewError(internalError error, outputMessage string, httpCode int) *Error {
	return &Error {
		InternalError: internalError,
		ErrorMessage: outputMessage,
		HttpCode: httpCode,
	}
}

func InternalServerError(err error) *Error {
	return &Error{
		InternalError: err,
		ErrorMessage: "internal server error",
		HttpCode: http.StatusInternalServerError,
	}
}

func WriteError(w http.ResponseWriter, err *Error){
	w.Header().Set("Content-Type", "application/json")

	log.Printf("[ERR] %s", err.InternalError)

	var js struct {
		Error string `json:"error"`
		Code int `json:"code"`
	}

	js.Error = err.ErrorMessage
	js.Code = err.HttpCode

	w.WriteHeader(err.HttpCode)
	json.NewEncoder(w).Encode(js)
}

