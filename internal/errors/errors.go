package errors

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
