package valid

// Numeric check number
type Numeric struct {
	Key string
}

// IsSatisfied judge whether obj is valid
func (n Numeric) IsSatisfied(obj interface{}) bool {
	if str, ok := obj.(string); ok {
		for _, v := range str {
			if '9' < v || v < '0' {
				return false
			}
		}
		return true
	}
	return false
}

// DefaultMessage return the default Length error message
func (n Numeric) DefaultMessage() string {
	return MessageTmpfs["Numeric"]
}

// GetKey return the n.Key
func (n Numeric) GetKey() string {
	return n.Key
}

// GetLimitValue return the limit value
func (n Numeric) GetLimitValue() interface{} {
	return nil
}
