package errors

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Error struct {
	InternalError error
	ErrorMessage  string
	OutputMessage string
	HttpCode      int
}

func NewError(internalError error, errMessage string, outputMessage string, httpCode int) *Error {
	return &Error{
		InternalError: internalError,
		ErrorMessage:  errMessage,
		OutputMessage: outputMessage,
		HttpCode:      httpCode,
	}
}

func InternalServerError(internalError error, errMessage string) *Error {
	return &Error{
		InternalError: internalError,
		ErrorMessage:  errMessage,
		OutputMessage: "internal server error",
		HttpCode:      http.StatusInternalServerError,
	}
}

func WriteError(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/json")

	slog.Error(
		err.ErrorMessage,
		"error", err.InternalError,
		"ReturningHTTPCode", err.HttpCode,
	)

	var js struct {
		Error string `json:"error"`
		Code  int    `json:"code"`
	}

	js.Error = err.OutputMessage
	js.Code = err.HttpCode

	w.WriteHeader(err.HttpCode)
	json.NewEncoder(w).Encode(js)
}
