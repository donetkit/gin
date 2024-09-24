package valid

// AlphaNumeric check alpha and number
type AlphaNumeric struct {
	Key string
}

// IsSatisfied judge whether obj is valid
func (a AlphaNumeric) IsSatisfied(obj interface{}) bool {
	if str, ok := obj.(string); ok {
		for _, v := range str {
			if ('Z' < v || v < 'A') && ('z' < v || v < 'a') && ('9' < v || v < '0') {
				return false
			}
		}
		return true
	}
	return false
}

// DefaultMessage return the default Length error message
func (a AlphaNumeric) DefaultMessage() string {
	return MessageTmpfs["AlphaNumeric"]
}

// GetKey return the a.Key
func (a AlphaNumeric) GetKey() string {
	return a.Key
}

// GetLimitValue return the limit value
func (a AlphaNumeric) GetLimitValue() interface{} {
	return nil
}
