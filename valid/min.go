package valid

import "fmt"

// Min check struct
type Min struct {
	Min int
	Key string
}

// IsSatisfied judge whether obj is valid
// not support int64 on 32-bit platform
func (m Min) IsSatisfied(obj interface{}) bool {
	var v int
	switch obj.(type) {
	case int64:
		if wordSize == 32 {
			return false
		}
		v = int(obj.(int64))
	case int:
		v = obj.(int)
	case int32:
		v = int(obj.(int32))
	case int16:
		v = int(obj.(int16))
	case int8:
		v = int(obj.(int8))
	default:
		return false
	}

	return v >= m.Min
}

// DefaultMessage return the default min error message
func (m Min) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["Min"], m.Min)
}

// GetKey return the m.Key
func (m Min) GetKey() string {
	return m.Key
}

// GetLimitValue return the limit value, Min
func (m Min) GetLimitValue() interface{} {
	return m.Min
}
