package lib

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func ResponseBadRequest(c *gin.Context, err error, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"message": message,
		"error":   err.Error()})
}
