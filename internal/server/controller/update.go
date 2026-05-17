package controller

import (
	"net/http"
	"strings"

	"slimebot/internal/updater"
)

func (h *HTTPController) GetUpdateCheck(c WebContext) {
	if h.update == nil {
		jsonError(c, http.StatusServiceUnavailable, "Update service is not configured.")
		return
	}
	force := strings.TrimSpace(c.Request().URL.Query().Get("force")) == "1"
	result, err := h.update.Check(c.Request().Context(), force)
	if err != nil {
		jsonInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPController) GetUpdateJob(c WebContext) {
	if h.update == nil {
		jsonError(c, http.StatusServiceUnavailable, "Update service is not configured.")
		return
	}
	status, err := h.update.Status(c.Request().Context())
	if err != nil {
		jsonInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *HTTPController) ApplyUpdate(c WebContext) {
	if h.update == nil {
		jsonError(c, http.StatusServiceUnavailable, "Update service is not configured.")
		return
	}
	var req updater.ApplyRequest
	if !bindJSONOrBadRequest(c, &req, "Invalid request payload format.") {
		return
	}
	status, err := h.update.Apply(c.Request().Context(), req)
	if err != nil {
		jsonInternalError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, status)
}
