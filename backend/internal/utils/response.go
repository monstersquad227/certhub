package utils

import "github.com/gin-gonic/gin"

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, httpCode, code int, message string) {
	c.JSON(httpCode, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}


