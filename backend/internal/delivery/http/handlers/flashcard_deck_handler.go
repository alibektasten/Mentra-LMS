package handlers

import (
	"net/http"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FlashcardDeckHandler struct {
	deckUC *usecases.FlashcardDeckUseCase
}

func NewFlashcardDeckHandler(deckUC *usecases.FlashcardDeckUseCase) *FlashcardDeckHandler {
	return &FlashcardDeckHandler{deckUC: deckUC}
}

func (h *FlashcardDeckHandler) Create(c *gin.Context) {
	var req dto.CreateFlashcardDeckRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	deck, err := h.deckUC.Create(c.Request.Context(), userID, orgID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": deck})
}

func (h *FlashcardDeckHandler) List(c *gin.Context) {
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	decks, err := h.deckUC.List(c.Request.Context(), userID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": decks})
}

func (h *FlashcardDeckHandler) Review(c *gin.Context) {
	deckID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid deck id"))
		return
	}
	var req dto.ReviewFlashcardRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	deck, err := h.deckUC.Review(c.Request.Context(), userID, orgID, deckID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": deck})
}

func (h *FlashcardDeckHandler) Delete(c *gin.Context) {
	deckID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid deck id"))
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	if err := h.deckUC.Delete(c.Request.Context(), userID, orgID, deckID); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
