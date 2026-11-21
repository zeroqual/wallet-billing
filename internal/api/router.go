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
	v1 := r.Group("/v1/accounts")
	v1.POST("", h.Create)
	v1.GET("/:id", h.Get)
	v1.POST("/deposit", h.Deposit)
	v1.POST("/withdraw", h.Withdraw)
	return r
}
