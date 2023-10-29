package api

type err struct {
	code string
	Msg  string
}

func (e *err) Error() string {
	return e.Msg
}

func NewError(code string, msg string) *err {
	return &err{
		code: code,
		Msg:  msg,
	}
}

func ConnCloseError() *err {
	return NewError(ErrorCode_ConnCloseError.String(), ErrorCode_ConnCloseError.String())
}

func IsConnCloseError(err err) bool {
	return err.code == ErrorCode_ConnCloseError.String()
}
