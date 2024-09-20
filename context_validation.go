package gin

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	// ValidTag struct tag
	ValidTag = "valid"

	LabelTag = "label"

	wordSize = 32 << (^uint(0) >> 32 & 1)
)

var (
	// key: function name
	// value: the number of parameters
	reflectFunc = make(ReflectFunc)

	// doesn't belong to validation functions
	unReflectFunc = map[string]bool{
		"Clear":     true,
		"HasErrors": true,
		"ErrorMap":  true,
		"Error":     true,
		"apply":     true,
		"Check":     true,
		"Valid":     true,
		"NoMatch":   true,
	}
	// ErrInt64On32 show 32-bit platform not support int64
	ErrInt64On32 = fmt.Errorf("not support int64 on 32-bit platform")
)

func init() {
	v := &Validation{}
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if !unReflectFunc[m.Name] {
			reflectFunc[m.Name] = m.Func
		}
	}
}

// CustomFunc is for custom validate function
type CustomFunc func(v *Validation, obj interface{}, key string)

// AddCustomFunc Add a custom function to validation
// The name can not be:
//
//	Clear
//	HasErrors
//	ErrorMap
//	Error
//	Check
//	Valid
//	NoMatch
//
// If the name is same with exists function, it will replace the origin valid function
func AddCustomFunc(name string, f CustomFunc) error {
	if unReflectFunc[name] {
		return fmt.Errorf("invalid function name: %s", name)
	}

	reflectFunc[name] = reflect.ValueOf(f)
	return nil
}

// ValidFunc Valid function type
type ValidFunc struct {
	Name   string
	Params []interface{}
}

// ReflectFunc Validate function map
type ReflectFunc map[string]reflect.Value

// Call validate values with named type string
func (f ReflectFunc) Call(name string, params ...interface{}) (result []reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	if _, ok := f[name]; !ok {
		err = fmt.Errorf("%s does not exist", name)
		return
	}
	if len(params) != f[name].Type().NumIn() {
		err = fmt.Errorf("the number of params is not adapted")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f[name].Call(in)
	return
}

func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func getValidFunc(f reflect.StructField) (vfs []ValidFunc, err error) {
	tag := f.Tag.Get(ValidTag)
	label := f.Tag.Get(LabelTag)
	if len(tag) == 0 {
		return
	}
	if vfs, tag, err = getRegFunc(tag, f.Name); err != nil {
		return
	}
	fs := strings.Split(tag, ";")
	for _, v := range fs {
		var vf ValidFunc
		if len(v) == 0 {
			continue
		}
		vf, err = parseFunc(v, f.Name, label)
		if err != nil {
			return
		}
		vfs = append(vfs, vf)
	}
	return
}

func getValidPFunc(reflectType reflect.Type) (vfs []ValidFunc, err error) {
	if reflectType.Kind() != reflect.Struct {
		return
	}
	validFunc := make([]ValidFunc, 0)
	for i := 0; i < reflectType.NumField(); i++ {
		tagJson := reflectType.Field(i).Tag.Get("json")
		if tagJson == "" {
			vfs, err = getValidPFunc(reflectType.Field(i).Type)
			for _, v := range vfs {
				validFunc = append(validFunc, v)
			}
			continue
		}
		tag := reflectType.Field(i).Tag.Get(ValidTag)
		label := reflectType.Field(i).Tag.Get(LabelTag)
		if len(tag) == 0 {
			vfs = validFunc
			return
		}
		if vfs, tag, err = getRegFunc(tag, reflectType.Field(i).Name); err != nil {
			vfs = validFunc
			return
		}
		fs := strings.Split(tag, ";")
		for _, v := range fs {
			var vf ValidFunc
			if len(v) == 0 {
				continue
			}
			vf, err = parseFunc(v, reflectType.Field(i).Name, label)
			if err != nil {
				vfs = validFunc
				return
			}
			validFunc = append(validFunc, vf)
		}
	}
	vfs = validFunc
	return
}

// Get Match function
// May be Get NoMatch function in the future
func getRegFunc(tag, key string) (vfs []ValidFunc, str string, err error) {
	tag = strings.TrimSpace(tag)
	index := strings.Index(tag, "Match(/")
	if index == -1 {
		str = tag
		return
	}
	end := strings.LastIndex(tag, "/)")
	if end < index {
		err = fmt.Errorf("invalid Match function")
		return
	}
	reg, err := regexp.Compile(tag[index+len("Match(/") : end])
	if err != nil {
		return
	}
	vfs = []ValidFunc{{"Match", []interface{}{reg, key + ".Match"}}}
	str = strings.TrimSpace(tag[:index]) + strings.TrimSpace(tag[end+len("/)"):])
	return
}

func parseFunc(vFunc, key string, label string) (v ValidFunc, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	vFunc = strings.TrimSpace(vFunc)
	start := strings.Index(vFunc, "(")
	var num int

	// doesn't need parameter valid function
	if start == -1 {
		if num, err = numIn(vFunc); err != nil {
			return
		}
		if num != 0 {
			err = fmt.Errorf("%s require %d parameters", vFunc, num)
			return
		}
		v = ValidFunc{vFunc, []interface{}{key + "." + vFunc + "." + label}}
		return
	}

	end := strings.Index(vFunc, ")")
	if end == -1 {
		err = fmt.Errorf("invalid valid function")
		return
	}

	name := strings.TrimSpace(vFunc[:start])
	if num, err = numIn(name); err != nil {
		return
	}

	params := strings.Split(vFunc[start+1:end], ",")
	// the num of param must be equal
	if num != len(params) {
		err = fmt.Errorf("%s require %d parameters", name, num)
		return
	}

	tParams, err := trim(name, key+"."+name+"."+label, params)
	if err != nil {
		return
	}
	v = ValidFunc{name, tParams}
	return
}

func numIn(name string) (num int, err error) {
	fn, ok := reflectFunc[name]
	if !ok {
		err = fmt.Errorf("doesn't exists %s valid function", name)
		return
	}
	// sub *Validation obj and key
	num = fn.Type().NumIn() - 3
	return
}

func trim(name, key string, s []string) (ts []interface{}, err error) {
	ts = make([]interface{}, len(s), len(s)+1)
	fn, ok := reflectFunc[name]
	if !ok {
		err = fmt.Errorf("doesn't exists %s valid function", name)
		return
	}
	for i := 0; i < len(s); i++ {
		var param interface{}
		// skip *Validation and obj params
		if param, err = parseParam(fn.Type().In(i+2), strings.TrimSpace(s[i])); err != nil {
			return
		}
		ts[i] = param
	}
	ts = append(ts, key)
	return
}

// modify the Parameters type to adapt the function input parameters' type
func parseParam(t reflect.Type, s string) (i interface{}, err error) {
	switch t.Kind() {
	case reflect.Int:
		i, err = strconv.Atoi(s)
	case reflect.Int64:
		if wordSize == 32 {
			return nil, ErrInt64On32
		}
		i, err = strconv.ParseInt(s, 10, 64)
	case reflect.Int32:
		var v int64
		v, err = strconv.ParseInt(s, 10, 32)
		if err == nil {
			i = int32(v)
		}
	case reflect.Int16:
		var v int64
		v, err = strconv.ParseInt(s, 10, 16)
		if err == nil {
			i = int16(v)
		}
	case reflect.Int8:
		var v int64
		v, err = strconv.ParseInt(s, 10, 8)
		if err == nil {
			i = int8(v)
		}
	case reflect.String:
		i = s
	case reflect.Ptr:
		if t.Elem().String() != "regexp.Regexp" {
			err = fmt.Errorf("not support %s", t.Elem().String())
			return
		}
		i, err = regexp.Compile(s)
	default:
		err = fmt.Errorf("not support %s", t.Kind().String())
	}
	return
}

func mergeParam(v *Validation, obj interface{}, params []interface{}) []interface{} {
	return append([]interface{}{v, obj}, params...)
}

// ======

// ValidFormer valid interface
type ValidFormer interface {
	Valid(*Validation)
}

// ErrorValidation show the error
type ErrorValidation struct {
	Message, Key, Name, Field, Tmpl string
	Value                           interface{}
	LimitValue                      interface{}
}

// String Returns the Message.
func (e *ErrorValidation) String() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// Implement Error interface.
// Return e.String()
func (e *ErrorValidation) Error() string { return e.String() }

// Result is returned from every validation method.
// It provides an indication of success, and a pointer to the Error (if any).
type Result struct {
	Error *ErrorValidation
	Ok    bool
}

// Key Get Result by given key string.
func (r *Result) Key(key string) *Result {
	if r.Error != nil {
		r.Error.Key = key
	}
	return r
}

// Message SetCalendar Result message by string or format string with args
func (r *Result) Message(message string, args ...interface{}) *Result {
	if r.Error != nil {
		if len(args) == 0 {
			r.Error.Message = message
		} else {
			r.Error.Message = fmt.Sprintf(message, args...)
		}
	}
	return r
}

// A Validation context manages data validation and error messages.
type Validation struct {
	// if this field set true, in struct tag valid
	// if the struct field vale is empty
	// it will skip those valid functions, see CanSkipFunc
	RequiredFirst bool

	Errors    []*ErrorValidation
	ErrorsMap map[string][]*ErrorValidation
}

// Clear Clean all ValidationError.
func (v *Validation) Clear() {
	v.Errors = []*ErrorValidation{}
	v.ErrorsMap = nil
}

// HasErrors Has ValidationError nor not.
func (v *Validation) HasErrors() bool {
	return len(v.Errors) > 0
}

// ErrorMap Return the errors mapped by key.
// If there are multiple validation errors associated with a single key, the
// first one "wins".  (Typically the first validation will be the more basic).
func (v *Validation) ErrorMap() map[string][]*ErrorValidation {
	return v.ErrorsMap
}

// Error Add an error to the validation context.
func (v *Validation) Error(message string, args ...interface{}) *Result {
	result := (&Result{
		Ok:    false,
		Error: &ErrorValidation{},
	}).Message(message, args...)
	v.Errors = append(v.Errors, result.Error)
	return result
}

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

func (v *Validation) apply(chk Validator, obj interface{}) *Result {
	if nil == obj {
		if chk.IsSatisfied(obj) {
			return &Result{Ok: true}
		}
	} else if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		if reflect.ValueOf(obj).IsNil() {
			if chk.IsSatisfied(nil) {
				return &Result{Ok: true}
			}
		} else {
			if chk.IsSatisfied(reflect.ValueOf(obj).Elem().Interface()) {
				return &Result{Ok: true}
			}
		}
	} else if chk.IsSatisfied(obj) {
		return &Result{Ok: true}
	}

	// Add the error to the validation context.
	key := chk.GetKey()
	Name := key
	Field := ""
	Label := ""
	parts := strings.Split(key, ".")
	if len(parts) == 3 {
		Field = parts[0]
		Name = parts[1]
		Label = parts[2]
		if len(Label) == 0 {
			Label = Field
		}
	}

	err := &ErrorValidation{
		Message:    Label + " " + chk.DefaultMessage(),
		Key:        key,
		Name:       Name,
		Field:      Field,
		Value:      obj,
		Tmpl:       MessageTmpfs[Name],
		LimitValue: chk.GetLimitValue(),
	}
	v.setError(err)

	// Also return it in the result.
	return &Result{
		Ok:    false,
		Error: err,
	}
}

// AddError key must like aa.bb.cc or aa.bb.
// AddError adds independent error message for the provided key
func (v *Validation) AddError(key, message string) {
	Name := key
	Field := ""

	Label := ""
	parts := strings.Split(key, ".")
	if len(parts) == 3 {
		Field = parts[0]
		Name = parts[1]
		Label = parts[2]
		if len(Label) == 0 {
			Label = Field
		}
	}

	err := &ErrorValidation{
		Message: Label + " " + message,
		Key:     key,
		Name:    Name,
		Field:   Field,
	}
	v.setError(err)
}

func (v *Validation) setError(err *ErrorValidation) {
	v.Errors = append(v.Errors, err)
	if v.ErrorsMap == nil {
		v.ErrorsMap = make(map[string][]*ErrorValidation)
	}
	if _, ok := v.ErrorsMap[err.Field]; !ok {
		v.ErrorsMap[err.Field] = []*ErrorValidation{}
	}
	v.ErrorsMap[err.Field] = append(v.ErrorsMap[err.Field], err)
}

// SetError SetCalendar error message for one field in ValidationError
func (v *Validation) SetError(fieldName string, errMsg string) *ErrorValidation {
	err := &ErrorValidation{Key: fieldName, Field: fieldName, Tmpl: errMsg, Message: errMsg}
	v.setError(err)
	return err
}

// Check Apply a group of validators to a field, in order, and return the
// ValidationResult from the first one that fails, or the last one that
// succeeds.
func (v *Validation) Check(obj interface{}, checks ...Validator) *Result {
	var result *Result
	for _, check := range checks {
		result = v.apply(check, obj)
		if !result.Ok {
			return result
		}
	}
	return result
}

//
//// 层级递增解析tag
//func GetReflectTag(reflectType reflect.Type, vfs *[]ValidFunc) {
//	if reflectType.Kind() != reflect.Struct {
//		return
//	}
//	for i := 0; i < reflectType.NumField(); i++ {
//		tag := reflectType.Field(i).Tag.Get("json")
//		if tag == "" {
//			GetReflectTag(reflectType.Field(i).Type, buf)
//			continue
//		}
//		buf.WriteString("`")
//		buf.WriteString(tag)
//		buf.WriteString("`,")
//	}
//}
//
//// 根据model中表模型的json标签获取表字段
//// 将select* 变为对应的字段名
//func GetColSQL(model interface{}) (sql string) {
//	var buf bytes.Buffer
//
//	typ := reflect.TypeOf(model)
//	for i := 0; i < typ.NumField(); i++ {
//		tag := typ.Field(i).Tag.Get("json")
//		if tag == "" {
//			GetReflectTag(typ.Field(i).Type, &buf)
//			continue
//		}
//		// sql += "`" + tag + "`,"
//		buf.WriteString("`")
//		buf.WriteString(tag)
//		buf.WriteString("`,")
//	}
//	sql = string(buf.Bytes()[:buf.Len()-1]) //去掉点,
//	return sql
//}

// Valid Validate a struct.
// the obj parameter must be a struct or a struct pointer
func (v *Validation) Valid(obj interface{}) (b bool, err error) {
	objT := reflect.TypeOf(obj)
	objV := reflect.ValueOf(obj)
	switch {
	case isStruct(objT):
	case isStructPtr(objT):
		objT = objT.Elem()
		objV = objV.Elem()
	default:
		err = fmt.Errorf("%v must be a struct or a struct pointer", obj)
		return
	}

	for i := 0; i < objT.NumField(); i++ {
		var vfs []ValidFunc
		tag := objT.Field(i).Tag.Get("json")
		if tag == "" {
			if vfs, err = getValidPFunc(objT.Field(i).Type); err != nil {
				return
			}
		} else {
			if vfs, err = getValidFunc(objT.Field(i)); err != nil {
				return
			}
		}
		//keys := objT.Field(i).Tag.Get("json")
		//fmt.Println(keys)
		var hasRequired bool
		for _, vf := range vfs {
			if vf.Name == "Required" {
				hasRequired = true
			}

			currentField := objV.Field(i).Interface()
			if objV.Field(i).Kind() == reflect.Ptr {
				if objV.Field(i).IsNil() {
					currentField = ""
				} else {
					currentField = objV.Field(i).Elem().Interface()
				}
			}

			chk := Required{}.IsSatisfied(currentField)
			if !hasRequired && v.RequiredFirst && !chk {
				if _, ok := CanSkipFunc[vf.Name]; ok {
					continue
				}
			}

			if _, err = reflectFunc.Call(vf.Name,
				mergeParam(v, objV.Field(i).Interface(), vf.Params)...); err != nil {
				return
			}
		}
	}

	if !v.HasErrors() {
		if form, ok := obj.(ValidFormer); ok {
			form.Valid(v)
		}
	}

	return !v.HasErrors(), nil
}

// RecursiveValid Recursively validate a struct.
// Step1: Validate by v.Valid
// Step2: If pass on step1, then reflect obj 's fields
// Step3: Do the Recursively validation to all struct or struct pointer fields
func (v *Validation) RecursiveValid(objc interface{}) (bool, error) {
	//Step 1: validate obj itself firstly
	// fails if objc is not struct
	pass, err := v.Valid(objc)
	if err != nil || !pass {
		return pass, err // Stop recursive validation
	}
	// Step 2: Validate struct s struct fields
	objT := reflect.TypeOf(objc)
	objV := reflect.ValueOf(objc)

	if isStructPtr(objT) {
		objT = objT.Elem()
		objV = objV.Elem()
	}

	for i := 0; i < objT.NumField(); i++ {

		t := objT.Field(i).Type

		// Recursive applies to struct or pointer to structs fields
		if isStruct(t) || isStructPtr(t) {
			// Step 3: do the recursive validation
			// Only valid the Public field recursively
			if objV.Field(i).CanInterface() {
				pass, err = v.RecursiveValid(objV.Field(i).Interface())
			}
		}
	}
	return pass, err
}

func (v *Validation) CanSkipAlso(skipFunc string) {
	if _, ok := CanSkipFunc[skipFunc]; !ok {
		CanSkipFunc[skipFunc] = struct{}{}
	}
}

// ======

// CanSkipFunc will skip valid if RequiredFirst is true and the struct field's value is empty
var CanSkipFunc = map[string]struct{}{
	"Email":   {},
	"IP":      {},
	"Mobile":  {},
	"Tel":     {},
	"Phone":   {},
	"ZipCode": {},
}

// MessageTmpfs store command validate template
var MessageTmpfs = map[string]string{
	"Required":     "不能为空",
	"Min":          "最小为 %d",
	"Max":          "最大为 %d",
	"Range":        "范围在 %d 至 %d",
	"MinSize":      "最小长度为 %d",
	"MaxSize":      "最大长度为 %d",
	"Length":       "长度必须是 %d",
	"Alpha":        "必须是有效的字母字符",
	"Numeric":      "必须是有效的数字字符",
	"AlphaNumeric": "必须是有效的字母或数字字符",
	"Match":        "必须匹配格式 %s",
	"NoMatch":      "必须不匹配格式 %s",
	"AlphaDash":    "必须是有效的字母或数字或破折号(-_)字符",
	"Email":        "必须是有效的邮件地址",
	"IP":           "必须是有效的IP地址",
	"Base64":       "必须是有效的base64字符",
	"Mobile":       "必须是有效手机号码",
	"Tel":          "必须是有效电话号码",
	"Phone":        "必须是有效的电话号码或者手机号码",
	"ZipCode":      "必须是有效的邮政编码",
	//"Required":     "Can not be empty",
	//"Min":          "Minimum is %d",
	//"Max":          "Maximum is %d",
	//"Range":        "Range is %d to %d",
	//"MinSize":      "Minimum size is %d",
	//"MaxSize":      "Maximum size is %d",
	//"Length":       "Required length is %d",
	//"Alpha":        "Must be valid alpha characters",
	//"Numeric":      "Must be valid numeric characters",
	//"AlphaNumeric": "Must be valid alpha or numeric characters",
	//"Match":        "Must match %s",
	//"NoMatch":      "Must not match %s",
	//"AlphaDash":    "Must be valid alpha or numeric or dash(-_) characters",
	//"Email":        "Must be a valid email address",
	//"IP":           "Must be a valid ip address",
	//"Base64":       "Must be valid base64 characters",
	//"Mobile":       "Must be valid mobile number",
	//"Tel":          "Must be valid telephone number",
	//"Phone":        "Must be valid telephone or mobile phone number",
	//"ZipCode":      "Must be valid zipcode",
}

var once sync.Once

// SetDefaultMessage set default messages
// if not set, the default messages are
//
//	"Required":     "Can not be empty",
//	"Min":          "Minimum is %d",
//	"Max":          "Maximum is %d",
//	"Range":        "Range is %d to %d",
//	"MinSize":      "Minimum size is %d",
//	"MaxSize":      "Maximum size is %d",
//	"Length":       "Required length is %d",
//	"Alpha":        "Must be valid alpha characters",
//	"Numeric":      "Must be valid numeric characters",
//	"AlphaNumeric": "Must be valid alpha or numeric characters",
//	"Match":        "Must match %s",
//	"NoMatch":      "Must not match %s",
//	"AlphaDash":    "Must be valid alpha or numeric or dash(-_) characters",
//	"Email":        "Must be a valid email address",
//	"IP":           "Must be a valid ip address",
//	"Base64":       "Must be valid base64 characters",
//	"Mobile":       "Must be valid mobile number",
//	"Tel":          "Must be valid telephone number",
//	"Phone":        "Must be valid telephone or mobile phone number",
//	"ZipCode":      "Must be valid zipcode",
func SetDefaultMessage(msg map[string]string) {
	if len(msg) == 0 {
		return
	}

	once.Do(func() {
		for name := range msg {
			MessageTmpfs[name] = msg[name]
		}
	})
	//fmt.Println(`you must SetDefaultMessage at once`)
}

// Validator interface
type Validator interface {
	IsSatisfied(interface{}) bool
	DefaultMessage() string
	GetKey() string
	GetLimitValue() interface{}
}

// Required struct
type Required struct {
	Key string
}

// IsSatisfied judge whether obj has value
func (r Required) IsSatisfied(obj interface{}) bool {
	if obj == nil {
		return false
	}

	if str, ok := obj.(string); ok {
		return len(strings.TrimSpace(str)) > 0
	}
	if _, ok := obj.(bool); ok {
		return true
	}
	if i, ok := obj.(int); ok {
		return i != 0
	}
	if i, ok := obj.(uint); ok {
		return i != 0
	}
	if i, ok := obj.(int8); ok {
		return i != 0
	}
	if i, ok := obj.(uint8); ok {
		return i != 0
	}
	if i, ok := obj.(int16); ok {
		return i != 0
	}
	if i, ok := obj.(uint16); ok {
		return i != 0
	}
	if i, ok := obj.(uint32); ok {
		return i != 0
	}
	if i, ok := obj.(int32); ok {
		return i != 0
	}
	if i, ok := obj.(int64); ok {
		return i != 0
	}
	if i, ok := obj.(uint64); ok {
		return i != 0
	}
	if t, ok := obj.(time.Time); ok {
		return !t.IsZero()
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Slice {
		return v.Len() > 0
	}
	return true
}

// DefaultMessage return the default error message
func (r Required) DefaultMessage() string {
	return MessageTmpfs["Required"]
}

// GetKey return the r.Key
func (r Required) GetKey() string {
	return r.Key
}

// GetLimitValue return nil now
func (r Required) GetLimitValue() interface{} {
	return nil
}

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

// Range Requires an integer to be within Min, Max inclusive.
type Range struct {
	Min
	Max
	Key string
}

// IsSatisfied judge whether obj is valid
// not support int64 on 32-bit platform
func (r Range) IsSatisfied(obj interface{}) bool {
	return r.Min.IsSatisfied(obj) && r.Max.IsSatisfied(obj)
}

// DefaultMessage return the default Range error message
func (r Range) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["Range"], r.Min.Min, r.Max.Max)
}

// GetKey return the m.Key
func (r Range) GetKey() string {
	return r.Key
}

// GetLimitValue return the limit value, Max
func (r Range) GetLimitValue() interface{} {
	return []int{r.Min.Min, r.Max.Max}
}

// MinSize Requires an array or string to be at least a given length.
type MinSize struct {
	Min int
	Key string
}

// IsSatisfied judge whether obj is valid
func (m MinSize) IsSatisfied(obj interface{}) bool {
	if str, ok := obj.(string); ok {
		return utf8.RuneCountInString(str) >= m.Min
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Slice {
		return v.Len() >= m.Min
	}
	return false
}

// DefaultMessage return the default MinSize error message
func (m MinSize) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["MinSize"], m.Min)
}

// GetKey return the m.Key
func (m MinSize) GetKey() string {
	return m.Key
}

// GetLimitValue return the limit value
func (m MinSize) GetLimitValue() interface{} {
	return m.Min
}

// MaxSize Requires an array or string to be at most a given length.
type MaxSize struct {
	Max int
	Key string
}

// IsSatisfied judge whether obj is valid
func (m MaxSize) IsSatisfied(obj interface{}) bool {
	if str, ok := obj.(string); ok {
		return utf8.RuneCountInString(str) <= m.Max
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Slice {
		return v.Len() <= m.Max
	}
	return false
}

// DefaultMessage return the default MaxSize error message
func (m MaxSize) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["MaxSize"], m.Max)
}

// GetKey return the m.Key
func (m MaxSize) GetKey() string {
	return m.Key
}

// GetLimitValue return the limit value
func (m MaxSize) GetLimitValue() interface{} {
	return m.Max
}

// Length Requires an array or string to be exactly a given length.
type Length struct {
	N   int
	Key string
}

// IsSatisfied judge whether obj is valid
func (l Length) IsSatisfied(obj interface{}) bool {
	if str, ok := obj.(string); ok {
		return utf8.RuneCountInString(str) == l.N
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Slice {
		return v.Len() == l.N
	}
	return false
}

// DefaultMessage return the default Length error message
func (l Length) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["Length"], l.N)
}

// GetKey return the m.Key
func (l Length) GetKey() string {
	return l.Key
}

// GetLimitValue return the limit value
func (l Length) GetLimitValue() interface{} {
	return l.N
}

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

// Match Requires a string to match a given regex.
type Match struct {
	Regexp *regexp.Regexp
	Key    string
}

// IsSatisfied judge whether obj is valid
func (m Match) IsSatisfied(obj interface{}) bool {
	return m.Regexp.MatchString(fmt.Sprintf("%v", obj))
}

// DefaultMessage return the default Match error message
func (m Match) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["Match"], m.Regexp.String())
}

// GetKey return the m.Key
func (m Match) GetKey() string {
	return m.Key
}

// GetLimitValue return the limit value
func (m Match) GetLimitValue() interface{} {
	return m.Regexp.String()
}

// NoMatch Requires a string to not match a given regex.
type NoMatch struct {
	Match
	Key string
}

// IsSatisfied judge whether obj is valid
func (n NoMatch) IsSatisfied(obj interface{}) bool {
	return !n.Match.IsSatisfied(obj)
}

// DefaultMessage return the default NoMatch error message
func (n NoMatch) DefaultMessage() string {
	return fmt.Sprintf(MessageTmpfs["NoMatch"], n.Regexp.String())
}

// GetKey return the n.Key
func (n NoMatch) GetKey() string {
	return n.Key
}

// GetLimitValue return the limit value
func (n NoMatch) GetLimitValue() interface{} {
	return n.Regexp.String()
}

var alphaDashPattern = regexp.MustCompile(`[^\d\w-_]`)

// AlphaDash check not Alpha
type AlphaDash struct {
	NoMatch
	Key string
}

// DefaultMessage return the default AlphaDash error message
func (a AlphaDash) DefaultMessage() string {
	return MessageTmpfs["AlphaDash"]
}

// GetKey return the n.Key
func (a AlphaDash) GetKey() string {
	return a.Key
}

// GetLimitValue return the limit value
func (a AlphaDash) GetLimitValue() interface{} {
	return nil
}

var emailPattern = regexp.MustCompile(`^[\w!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[\w!#$%&'*+/=?^_` + "`" + `{|}~-]+)*@(?:[\w](?:[\w-]*[\w])?\.)+[a-zA-Z0-9](?:[\w-]*[\w])?$`)

// Email check struct
type Email struct {
	Match
	Key string
}

// DefaultMessage return the default Email error message
func (e Email) DefaultMessage() string {
	return MessageTmpfs["Email"]
}

// GetKey return the n.Key
func (e Email) GetKey() string {
	return e.Key
}

// GetLimitValue return the limit value
func (e Email) GetLimitValue() interface{} {
	return nil
}

var ipPattern = regexp.MustCompile(`^((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`)

// IP check struct
type IP struct {
	Match
	Key string
}

// DefaultMessage return the default IP error message
func (i IP) DefaultMessage() string {
	return MessageTmpfs["IP"]
}

// GetKey return the i.Key
func (i IP) GetKey() string {
	return i.Key
}

// GetLimitValue return the limit value
func (i IP) GetLimitValue() interface{} {
	return nil
}

var base64Pattern = regexp.MustCompile(`^(?:[A-Za-z0-99+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`)

// Base64 check struct
type Base64 struct {
	Match
	Key string
}

// DefaultMessage return the default Base64 error message
func (b Base64) DefaultMessage() string {
	return MessageTmpfs["Base64"]
}

// GetKey return the b.Key
func (b Base64) GetKey() string {
	return b.Key
}

// GetLimitValue return the limit value
func (b Base64) GetLimitValue() interface{} {
	return nil
}

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

// just for chinese telephone number
var telPattern = regexp.MustCompile(`^(0\d{2,3}(\-)?)?\d{7,8}$`)

// Tel check telephone struct
type Tel struct {
	Match
	Key string
}

// DefaultMessage return the default Tel error message
func (t Tel) DefaultMessage() string {
	return MessageTmpfs["Tel"]
}

// GetKey return the t.Key
func (t Tel) GetKey() string {
	return t.Key
}

// GetLimitValue return the limit value
func (t Tel) GetLimitValue() interface{} {
	return nil
}

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
