package response

import "github.com/gin-gonic/gin"

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Detail  any    `json:"detail,omitempty"`
}

func Error(c *gin.Context, status int, code int, message string, detail any) {
	c.JSON(status, Response{
		Status:  code,
		Message: message,
		Detail:  detail,
	})
}

func Success(c *gin.Context, data any) {
	c.JSON(200, data)
}

func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}
