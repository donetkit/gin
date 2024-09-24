package valid

import "fmt"

// Max validate struct
type Max struct {
	Max int
	Key string
}

// IsSatisfied judge whether obj is valid
// not support int64 on 32-bit platform
func (m Max) IsSatisfied(obj interface{}) bool {
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

	return v <= m.Max
}

// DefaultMessage return the default max error message
func (m Max) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["Max"], m.Max)
}

// GetKey return the m.Key
func (m Max) GetKey() string {
	return m.Key
}

// GetLimitValue return the limit value, Max
func (m Max) GetLimitValue() interface{} {
	return m.Max
}
