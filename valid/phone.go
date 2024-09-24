package valid

// Phone just for chinese telephone or mobile phone number
type Phone struct {
	Mobile
	Tel
	Key string
}

// IsSatisfied judge whether obj is valid
func (p Phone) IsSatisfied(obj interface{}) bool {
	return p.Mobile.IsSatisfied(obj) || p.Tel.IsSatisfied(obj)
}

// DefaultMessage return the default Phone error message
func (p Phone) DefaultMessage() string {
	return MessageTmpfs["Phone"]
}

// GetKey return the p.Key
func (p Phone) GetKey() string {
	return p.Key
}

// GetLimitValue return the limit value
func (p Phone) GetLimitValue() interface{} {
	return nil
}
