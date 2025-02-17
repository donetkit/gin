package valid

import (
	"net/url"
)

// Url Requires an array or string to be exactly a given length.
type Url struct {
	Key string
}

// IsSatisfied judge whether obj is valid
func (l Url) IsSatisfied(obj interface{}) bool {
	if str, ok := obj.(string); ok {
		return isValidURL(str)
	}
	return false
}

// DefaultMessage return the default Url error message
func (l Url) DefaultMessage() string {
	return MessageTmpfs["Url"]
}

// GetKey return the m.Key
func (l Url) GetKey() string {
	return l.Key
}

// GetLimitValue return the limit value
func (l Url) GetLimitValue() interface{} {
	return nil
}

func isValidURL(input string) bool {
	_, err := url.ParseRequestURI(input)
	if err != nil {
		return false
	}
	u, err := url.Parse(input)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}
