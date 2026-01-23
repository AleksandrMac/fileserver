package uerror

type UError interface {
	error
	Status() int
	Message() string
	Payload() map[string]any
}

type uerror struct {
	StatusF  int
	MessageF string
	ErrF     error
	PayloadF map[string]any
}

func NewUError(status int, message string, err error, payload map[string]any) UError {
	return uerror{status, message, err, payload}
}

func (x uerror) Status() int {
	return x.StatusF
}

func (x uerror) Message() string {
	return x.MessageF
}

func (x uerror) Error() string {
	return x.ErrF.Error()
}

func (x uerror) Payload() map[string]any {
	return x.PayloadF
}
