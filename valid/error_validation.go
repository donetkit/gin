package valid

// ErrorValidation show the error
type ErrorValidation struct {
	Message, Key, Name, Field, Tmpl string
	Value                           interface{}
	LimitValue                      interface{}
}

func (e *ErrorValidation) Error() string {
	return e.String()
}

// String Returns the Message.
func (e *ErrorValidation) String() string {
	if e == nil {
		return ""
	}
	return e.Message
}
