package valid

import (
	"reflect"
	"strings"
	"time"
)

// Required struct
type Required struct {
	Key string
}

// IsSatisfied judge whether obj has value
func (r Required) IsSatisfied(obj interface{}) bool {
	if obj == nil {
		return false
	}

	if str, ok := obj.(string); ok {
		return len(strings.TrimSpace(str)) > 0
	}
	if _, ok := obj.(bool); ok {
		return true
	}
	if i, ok := obj.(int); ok {
		return i != 0
	}
	if i, ok := obj.(uint); ok {
		return i != 0
	}
	if i, ok := obj.(int8); ok {
		return i != 0
	}
	if i, ok := obj.(uint8); ok {
		return i != 0
	}
	if i, ok := obj.(int16); ok {
		return i != 0
	}
	if i, ok := obj.(uint16); ok {
		return i != 0
	}
	if i, ok := obj.(uint32); ok {
		return i != 0
	}
	if i, ok := obj.(int32); ok {
		return i != 0
	}
	if i, ok := obj.(int64); ok {
		return i != 0
	}
	if i, ok := obj.(uint64); ok {
		return i != 0
	}
	if t, ok := obj.(time.Time); ok {
		return !t.IsZero()
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Slice {
		return v.Len() > 0
	}
	return true
}

// DefaultMessage return the default error message
func (r Required) DefaultMessage() string {
	return MessageTmpfs["Required"]
}

// GetKey return the r.Key
func (r Required) GetKey() string {
	return r.Key
}

// GetLimitValue return nil now
func (r Required) GetLimitValue() interface{} {
	return nil
}
