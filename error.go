package vermouth

type RuntimeError struct {
	Code    int
	Message string
}

func (e *RuntimeError) Error() string {
	return e.Message
}

func NewRuntimeError(code int, message string) *RuntimeError {
	return &RuntimeError{Code: code, Message: message}
}