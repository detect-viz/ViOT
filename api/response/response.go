package response

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是 API 響應的標準格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

// Success 返回成功響應
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	})
}

// Fail 返回失敗響應
func Fail(c *gin.Context, code int, message string, err string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
		Error:   err,
	})
}

// BadRequest 返回請求錯誤響應
func BadRequest(c *gin.Context, message string, err string) {
	Fail(c, http.StatusBadRequest, message, err)
}

// Unauthorized 返回未授權響應
func Unauthorized(c *gin.Context, message string, err string) {
	Fail(c, http.StatusUnauthorized, message, err)
}

// Forbidden 返回禁止訪問響應
func Forbidden(c *gin.Context, message string, err string) {
	Fail(c, http.StatusForbidden, message, err)
}

// NotFound 返回資源不存在響應
func NotFound(c *gin.Context, message string, err string) {
	Fail(c, http.StatusNotFound, message, err)
}

// InternalServerError 返回服務器錯誤響應
func InternalServerError(c *gin.Context, message string, err string) {
	Fail(c, http.StatusInternalServerError, message, err)
}

// respondJSON 響應 JSON 數據
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// respondError 響應錯誤
func RespondError(w http.ResponseWriter, code int, message string) {
	RespondJSON(w, code, map[string]string{"error": message})
}
