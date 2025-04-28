package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Ping
// @Summary      Ping API
// @Description  Checks if the service is up and running.
// @Tag         Health
// @Success      200  string  "pong"
// @Router       /api/ping [get]
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, "pong")
}
