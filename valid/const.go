package valid

import "sync"

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
	"Min":          "最小为%d",
	"Max":          "最大为%d",
	"Range":        "范围在%d至%d",
	"MinSize":      "最小为%d",
	"MaxSize":      "最大为%d",
	"Length":       "长度必须为%d",
	"Alpha":        "必须是有效的字母字符",
	"Numeric":      "必须是有效的数字字符",
	"AlphaNumeric": "必须是有效的字母或数字字符",
	"Match":        "必须匹配格式: %s",
	"NoMatch":      "必须不匹配格式: %s",
	"AlphaDash":    "必须是有效的字母或数字或破折号(-_)字符",
	"Email":        "必须是有效的邮件地址",
	"IP":           "必须是有效的IP地址",
	"Base64":       "必须是有效的base64字符",
	"Mobile":       "必须是有效手机号码",
	"Tel":          "必须是有效电话号码",
	"Phone":        "必须是有效的电话号码或者手机号码",
	"ZipCode":      "必须是有效的邮政编码",
	"Repeat":       "必须是不重复的数据",
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
