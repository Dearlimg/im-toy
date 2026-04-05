package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"im-chat/internal/apperrors"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: GetMsg(CodeSuccess),
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, data interface{}, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: msg,
		Data:    data,
	})
}

func Failed(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeFailed,
		Message: msg,
		Data:    nil,
	})
}

func FailedWithCode(c *gin.Context, code int) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: GetMsg(code),
		Data:    nil,
	})
}

func FailedWithCodeAndMsg(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
		Data:    nil,
	})
}

func InvalidParams(c *gin.Context, msg string) {
	if msg == "" {
		msg = GetMsg(CodeInvalidParams)
	}
	c.JSON(http.StatusOK, Response{
		Code:    CodeInvalidParams,
		Message: msg,
		Data:    nil,
	})
}

func Unauthorized(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeUnauthorized,
		Message: GetMsg(CodeUnauthorized),
		Data:    nil,
	})
}

func Forbidden(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeForbidden,
		Message: GetMsg(CodeForbidden),
		Data:    nil,
	})
}

func NotFound(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeNotFound,
		Message: GetMsg(CodeNotFound),
		Data:    nil,
	})
}

func InternalError(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeInternalError,
		Message: GetMsg(CodeInternalError),
		Data:    nil,
	})
}

func AppError(c *gin.Context, err *apperrors.AppError) {
	c.JSON(http.StatusOK, Response{
		Code:    err.Code,
		Message: err.Msg,
		Data:    nil,
	})
}
