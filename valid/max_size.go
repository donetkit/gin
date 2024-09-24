package valid

import (
	"fmt"
	"reflect"
	"unicode/utf8"
)

// MaxSize Requires an array or string to be at most a given length.
type MaxSize struct {
	Max int
	Key string
}

// IsSatisfied judge whether obj is valid
func (m MaxSize) IsSatisfied(obj interface{}) bool {
	if str, ok := obj.(string); ok {
		return utf8.RuneCountInString(str) <= m.Max
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Slice {
		return v.Len() <= m.Max
	}
	return false
}

// DefaultMessage return the default MaxSize error message
func (m MaxSize) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["MaxSize"], m.Max)
}

// GetKey return the m.Key
func (m MaxSize) GetKey() string {
	return m.Key
}

// GetLimitValue return the limit value
func (m MaxSize) GetLimitValue() interface{} {
	return m.Max
}
