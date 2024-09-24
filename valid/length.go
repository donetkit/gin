package valid

import (
	"fmt"
	"reflect"
	"unicode/utf8"
)

// Length Requires an array or string to be exactly a given length.
type Length struct {
	N   int
	Key string
}

// IsSatisfied judge whether obj is valid
func (l Length) IsSatisfied(obj interface{}) bool {
	if str, ok := obj.(string); ok {
		return utf8.RuneCountInString(str) == l.N
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Slice {
		return v.Len() == l.N
	}
	return false
}

// DefaultMessage return the default Length error message
func (l Length) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["Length"], l.N)
}

// GetKey return the m.Key
func (l Length) GetKey() string {
	return l.Key
}

// GetLimitValue return the limit value
func (l Length) GetLimitValue() interface{} {
	return l.N
}
