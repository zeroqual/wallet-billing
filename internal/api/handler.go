package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Readyz(c *gin.Context) {
	if err := h.service.Pinger(c.Request.Context()); err != nil {
		slog.Warn("db not ready", slog.Any("err", err))
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "bad"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
