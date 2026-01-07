package metrics

import (
	repo "app/http/repository/metrics"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{ repo *repo.Repository }

func NewHandler(r *repo.Repository) *Handler { return &Handler{repo: r} }

type clickReq struct {
	Slot int `json:"slot"`
}

// TrackAdClick tracks advertisement click
// @Summary Track ad click
// @Description Track advertisement click for slot (1 or 2)
// @Tags metrics
// @Accept json
// @Produce json
// @Param slot body clickReq false "Slot click payload"
// @Param slot query int false "Slot via query param"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /metrics/ad-click [post]
func (h *Handler) TrackAdClick(c *gin.Context) {
	var req clickReq
	if err := c.ShouldBindJSON(&req); err != nil {
		// allow query param fallback: /metrics/ad-click?slot=1
		if v := c.Query("slot"); v == "2" {
			req.Slot = 2
		} else {
			req.Slot = 1
		}
	}
	if req.Slot != 1 && req.Slot != 2 {
		req.Slot = 1
	}
	if err := h.repo.Increment(req.Slot); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c1, c2, _ := h.repo.GetTotals()
	c.JSON(http.StatusOK, gin.H{"ok": true, "ad_clicks_1": c1, "ad_clicks_2": c2})
}
