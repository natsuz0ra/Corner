package controller

import (
	"net/http"

	"slimebot/internal/domain"
)

type planService interface {
	ListPlans() ([]domain.Plan, error)
	GetPlan(planID string) (*domain.Plan, error)
	UpdatePlanStatus(planID, status string) (*domain.Plan, error)
	DeletePlan(planID string) error
}

// ListPlans returns all plans.
func (h *HTTPController) ListPlans(c WebContext) {
	plans, err := h.plans.ListPlans()
	if err != nil {
		jsonInternalError(c, err)
		return
	}
	if plans == nil {
		plans = []domain.Plan{}
	}
	c.JSON(http.StatusOK, plans)
}

// GetPlan returns a single plan by ID.
func (h *HTTPController) GetPlan(c WebContext) {
	id := c.Param("id")
	plan, err := h.plans.GetPlan(id)
	if err != nil {
		jsonError(c, http.StatusNotFound, "Plan not found.")
		return
	}
	c.JSON(http.StatusOK, plan)
}

// UpdatePlanStatus changes a plan's status (approve/reject).
func (h *HTTPController) UpdatePlanStatus(c WebContext) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status"`
	}
	if !bindJSONOrBadRequest(c, &req, "Invalid request payload format.") {
		return
	}
	plan, err := h.plans.UpdatePlanStatus(id, req.Status)
	if err != nil {
		jsonInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

// DeletePlan removes a plan.
func (h *HTTPController) DeletePlan(c WebContext) {
	id := c.Param("id")
	if err := h.plans.DeletePlan(id); err != nil {
		jsonInternalError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
