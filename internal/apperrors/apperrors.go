package apperrors

import "fmt"

type AppError struct {
	Code int
	Msg  string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("code=%d, msg=%s", e.Code, e.Msg)
}

func New(code int, msg string) *AppError {
	return &AppError{Code: code, Msg: msg}
}

func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{Code: e.Code, Msg: msg}
}

func FromError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}
