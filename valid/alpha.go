package valid

// Alpha check the alpha
type Alpha struct {
	Key string
}

// IsSatisfied judge whether obj is valid
func (a Alpha) IsSatisfied(obj interface{}) bool {
	if str, ok := obj.(string); ok {
		for _, v := range str {
			if ('Z' < v || v < 'A') && ('z' < v || v < 'a') {
				return false
			}
		}
		return true
	}
	return false
}

// DefaultMessage return the default Length error message
func (a Alpha) DefaultMessage() string {
	return MessageTmpfs["Alpha"]
}

// GetKey return the m.Key
func (a Alpha) GetKey() string {
	return a.Key
}

// GetLimitValue return the limit value
func (a Alpha) GetLimitValue() interface{} {
	return nil
}
