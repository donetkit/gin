package valid

import "regexp"

// just for chinese zipcode
var zipCodePattern = regexp.MustCompile(`^[1-9]\d{5}$`)

// ZipCode check the zip struct
type ZipCode struct {
	Match
	Key string
}

// DefaultMessage return the default Zip error message
func (z ZipCode) DefaultMessage() string {
	return MessageTmpfs["ZipCode"]
}

// GetKey return the z.Key
func (z ZipCode) GetKey() string {
	return z.Key
}

// GetLimitValue return the limit value
func (z ZipCode) GetLimitValue() interface{} {
	return nil
}
