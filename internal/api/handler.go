package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zeroqual/wallet-billing/internal/domain"
	"github.com/zeroqual/wallet-billing/internal/service"
)

type createAccountDTO struct {
	OwnerID string `json:"owner_id" binding:"required"`
}

type accountResponse struct {
	ID        uuid.UUID `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toResponse(a domain.Account) accountResponse {
	return accountResponse{
		ID:        a.ID,
		OwnerID:   a.OwnerID,
		Balance:   a.Balance,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

type Pinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	accounts *service.AccountService
	db       Pinger
}

func NewHandler(accounts *service.AccountService, db Pinger) *Handler {
	return &Handler{accounts: accounts, db: db}
}

func (h *Handler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		slog.Warn("db not ready", slog.Any("err", err))
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *Handler) Create(c *gin.Context) {
	var dto createAccountDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		slog.Error("failed to bind json", slog.Any("err", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	acc, err := h.accounts.Create(c.Request.Context(), dto.OwnerID)
	if err != nil {
		slog.Error("got error create handler", slog.Any("err", err))
		switch {
		case errors.Is(err, domain.ErrAlreadyExists) ||
			errors.Is(err, domain.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{"account": toResponse(acc)})
}

func (h *Handler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a valid uuid"})
		return
	}

	acc, err := h.accounts.Get(c.Request.Context(), id)
	if err != nil {
		slog.Error("got error get handler", slog.Any("err", err))
		switch {
		case errors.Is(err, domain.ErrNotFound) ||
			errors.Is(err, domain.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"account": toResponse(acc)})
}
