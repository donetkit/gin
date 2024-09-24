package valid

import "fmt"

// Result is returned from every validation method.
// It provides an indication of success, and a pointer to the Error (if any).
type Result struct {
	Error *ErrorValidation
	Ok    bool
}

// Key Get Result by given key string.
func (r *Result) Key(key string) *Result {
	if r.Error != nil {
		r.Error.Key = key
	}
	return r
}

// Message SetCalendar Result message by string or format string with args
func (r *Result) Message(message string, args ...interface{}) *Result {
	if r.Error != nil {
		if len(args) == 0 {
			r.Error.Message = message
		} else {
			r.Error.Message = fmt.Sprintf(message, args...)
		}
	}
	return r
}
