package apperror

type AppError struct {
	Code    ErrCode
	Message string
	Err     error `json:"-"`
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func (c ErrCode) New(msg string) *AppError {
	return &AppError{Code: c, Message: msg}
}

func (c ErrCode) Wrap(err error, msg string) *AppError {
	return &AppError{Code: c, Message: msg, Err: err}
}
