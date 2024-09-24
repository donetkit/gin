package valid

import (
	"fmt"
	"reflect"
	"unicode/utf8"
)

// MinSize Requires an array or string to be at least a given length.
type MinSize struct {
	Min int
	Key string
}

// IsSatisfied judge whether obj is valid
func (m MinSize) IsSatisfied(obj interface{}) bool {
	if str, ok := obj.(string); ok {
		return utf8.RuneCountInString(str) >= m.Min
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Slice {
		return v.Len() >= m.Min
	}
	return false
}

// DefaultMessage return the default MinSize error message
func (m MinSize) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["MinSize"], m.Min)
}

// GetKey return the m.Key
func (m MinSize) GetKey() string {
	return m.Key
}

// GetLimitValue return the limit value
func (m MinSize) GetLimitValue() interface{} {
	return m.Min
}
