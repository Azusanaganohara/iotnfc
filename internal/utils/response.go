package utils

import "github.com/gin-gonic/gin"

type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func ResponseOK(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ResponseError(c *gin.Context, status int, message string, errMsg string) {
	c.JSON(status, StandardResponse{
		Success: false,
		Message: message,
		Error:   errMsg,
	})
}
