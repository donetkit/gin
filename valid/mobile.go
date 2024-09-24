package valid

import "regexp"

// just for chinese mobile phone number
var mobilePattern = regexp.MustCompile(`^((\+86)|(86))?1([356789][0-9]|4[579]|6[67]|7[0135678]|9[189])[0-9]{8}$`)

// Mobile check struct
type Mobile struct {
	Match
	Key string
}

// DefaultMessage return the default Mobile error message
func (m Mobile) DefaultMessage() string {
	return MessageTmpfs["Mobile"]
}

// GetKey return the m.Key
func (m Mobile) GetKey() string {
	return m.Key
}

// GetLimitValue return the limit value
func (m Mobile) GetLimitValue() interface{} {
	return nil
}
