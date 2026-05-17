package controller

import "net/http"

func (h *HTTPController) GetAgentsInstructions(c WebContext) {
	if h.agents == nil {
		jsonError(c, http.StatusServiceUnavailable, "AGENTS instructions service is not configured.")
		return
	}
	file, err := h.agents.ReadGlobal(c.Request().Context())
	if err != nil {
		jsonInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, map[string]string{
		"content": file.Content,
		"path":    file.Path,
	})
}

func (h *HTTPController) UpdateAgentsInstructions(c WebContext) {
	if h.agents == nil {
		jsonError(c, http.StatusServiceUnavailable, "AGENTS instructions service is not configured.")
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	if !bindJSONOrBadRequest(c, &req, "Invalid request payload format.") {
		return
	}
	if err := h.agents.UpdateGlobal(c.Request().Context(), req.Content); err != nil {
		jsonInternalError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
