package api

import (
	"github.com/gin-gonic/gin"
)

func NewRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", h.Healthz)
	r.GET("/readyz", h.Readyz)
	return r
}
