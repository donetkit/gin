package gin

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/gin-gonic/gin/valid"
)

var MsgFlags = map[int]string{
	SUCCESS:       "请求成功",
	ERROR:         "请求异常",
	FAIL:          "请求失败",
	InvalidParams: "请求参数错误",
}

// GetMsg get error information based on Code
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[ERROR]
}

const (
	SUCCESS       = 0
	FAIL          = -1
	ERROR         = 500
	InvalidParams = 400
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PageList struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    *PageData `json:"data"`
}

type PageData struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"` // 总数
}

// Success 请求成功
func (c *Context) Success(data interface{}) {
	c.JSON(200, &Response{
		Code:    SUCCESS,
		Message: GetMsg(SUCCESS),
		Data:    data,
	})
}

// SuccessPage 请求成功
func (c *Context) SuccessPage(data interface{}, total int64) {
	c.JSON(200, &PageList{
		Code:    SUCCESS,
		Message: GetMsg(SUCCESS),
		Data:    &PageData{List: data, Total: total},
	})
}

// Fail 请求失败
func (c *Context) Fail(code int, messages ...string) {
	msg := GetMsg(code)
	if len(messages) > 0 {
		msg = messages[0]
	}
	c.JSON(200, &Response{
		Code:    code,
		Message: msg,
		Data:    "",
	})
	c.Abort()
}

// InvalidParams 请求参数错误
func (c *Context) InvalidParams(code int, messages ...string) {
	msg := GetMsg(InvalidParams)
	if len(messages) > 0 {
		msg = messages[0]
	}
	c.JSON(200, &Response{
		Code:    code,
		Message: msg,
		Data:    "",
	})
	c.Abort()
}

// Unauthorized 请求未授权
func (c *Context) Unauthorized(httpStatus, code int, messages ...string) {
	msg := GetMsg(FAIL)
	if len(messages) > 0 {
		msg = messages[0]
	}
	c.JSON(httpStatus, &Response{
		Code:    code,
		Message: msg,
		Data:    "",
	})
	c.Abort()
}

// Exception 请求失败
func (c *Context) Exception(messages ...string) {
	msg := GetMsg(ERROR)
	if len(messages) > 0 {
		msg = messages[0]
	}
	c.JSON(200, Response{
		Code:    ERROR,
		Message: msg,
		Data:    "",
	})
	c.Abort()
}

func (c *Context) ShouldBindWithValid(s interface{}) error {
	b := binding.Default(c.Request.Method, c.ContentType())
	_ = c.ShouldBindWith(s, b)
	validation := valid.Validation{}
	check, err := validation.Valid(s)
	if err != nil {
		return err
	}
	if !check {
		return validation.Errors[0]
	}
	return nil
}
