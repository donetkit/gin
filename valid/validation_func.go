package valid

import (
	"regexp"
)

// Required Test that the argument is non-nil and non-empty (if string or list)
func (v *Validation) Required(obj interface{}, key string) *Result {
	return v.apply(Required{key}, obj)
}

// Min Test that the obj is greater than min if obj 's type is int
func (v *Validation) Min(obj interface{}, min int, key string) *Result {
	return v.apply(Min{min, key}, obj)
}

// Max Test that the obj is less than max if obj 's type is int
func (v *Validation) Max(obj interface{}, max int, key string) *Result {
	return v.apply(Max{max, key}, obj)
}

// Range Test that the obj is between mni and max if obj 's type is int
func (v *Validation) Range(obj interface{}, min, max int, key string) *Result {
	return v.apply(Range{Min{Min: min}, Max{Max: max}, key}, obj)
}

// MinSize Test that the obj is longer than min size if type is string or slice
func (v *Validation) MinSize(obj interface{}, min int, key string) *Result {
	return v.apply(MinSize{min, key}, obj)
}

// MaxSize Test that the obj is shorter than max size if type is string or slice
func (v *Validation) MaxSize(obj interface{}, max int, key string) *Result {
	return v.apply(MaxSize{max, key}, obj)
}

// Length Test that the obj is same length to n if type is string or slice
func (v *Validation) Length(obj interface{}, n int, key string) *Result {
	return v.apply(Length{n, key}, obj)
}

// Alpha Test that the obj is [a-zA-Z] if type is string
func (v *Validation) Alpha(obj interface{}, key string) *Result {
	return v.apply(Alpha{key}, obj)
}

// Numeric Test that the obj is [0-9] if type is string
func (v *Validation) Numeric(obj interface{}, key string) *Result {
	return v.apply(Numeric{key}, obj)
}

// AlphaNumeric Test that the obj is [0-9a-zA-Z] if type is string
func (v *Validation) AlphaNumeric(obj interface{}, key string) *Result {
	return v.apply(AlphaNumeric{key}, obj)
}

// Match Test that the obj matches regexp if type is string
func (v *Validation) Match(obj interface{}, regex *regexp.Regexp, key string) *Result {
	return v.apply(Match{regex, key}, obj)
}

// NoMatch Test that the obj doesn't match regexp if type is string
func (v *Validation) NoMatch(obj interface{}, regex *regexp.Regexp, key string) *Result {
	return v.apply(NoMatch{Match{Regexp: regex}, key}, obj)
}

// AlphaDash Test that the obj is [0-9a-zA-Z_-] if type is string
func (v *Validation) AlphaDash(obj interface{}, key string) *Result {
	return v.apply(AlphaDash{NoMatch{Match: Match{Regexp: alphaDashPattern}}, key}, obj)
}

// Email Test that the obj is email address if type is string
func (v *Validation) Email(obj interface{}, key string) *Result {
	return v.apply(Email{Match{Regexp: emailPattern}, key}, obj)
}

// IP Test that the obj is IP address if type is string
func (v *Validation) IP(obj interface{}, key string) *Result {
	return v.apply(IP{Match{Regexp: ipPattern}, key}, obj)
}

// Base64 Test that the obj is base64 encoded if type is string
func (v *Validation) Base64(obj interface{}, key string) *Result {
	return v.apply(Base64{Match{Regexp: base64Pattern}, key}, obj)
}

// Mobile Test that the obj is chinese mobile number if type is string
func (v *Validation) Mobile(obj interface{}, key string) *Result {
	return v.apply(Mobile{Match{Regexp: mobilePattern}, key}, obj)
}

// Tel Test that the obj is chinese telephone number if type is string
func (v *Validation) Tel(obj interface{}, key string) *Result {
	return v.apply(Tel{Match{Regexp: telPattern}, key}, obj)
}

// Phone Test that the obj is chinese mobile or telephone number if type is string
func (v *Validation) Phone(obj interface{}, key string) *Result {
	return v.apply(Phone{Mobile{Match: Match{Regexp: mobilePattern}},
		Tel{Match: Match{Regexp: telPattern}}, key}, obj)
}

// ZipCode Test that the obj is chinese zip code if type is string
func (v *Validation) ZipCode(obj interface{}, key string) *Result {
	return v.apply(ZipCode{Match{Regexp: zipCodePattern}, key}, obj)
}

// Repeat Test that the obj is remove duplicates
func (v *Validation) Repeat(obj interface{}, key string) *Result {
	return v.apply(&Repeat{key}, obj)
}

// Url Test that the obj is URL if type is string
func (v *Validation) Url(obj interface{}, key string) *Result {
	return v.apply(Url{key}, obj)
}
