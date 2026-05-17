package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"slimebot/internal/updater"
)

type updateServiceStub struct {
	checkForce bool
	applyReq   updater.ApplyRequest
	check      updater.CheckResult
	status     updater.JobStatus
	apply      updater.JobStatus
}

func (s *updateServiceStub) Check(_ context.Context, force bool) (updater.CheckResult, error) {
	s.checkForce = force
	return s.check, nil
}

func (s *updateServiceStub) Status(_ context.Context) (updater.JobStatus, error) {
	return s.status, nil
}

func (s *updateServiceStub) Apply(_ context.Context, req updater.ApplyRequest) (updater.JobStatus, error) {
	s.applyReq = req
	return s.apply, nil
}

func TestGetUpdateCheckReturnsResult(t *testing.T) {
	stub := &updateServiceStub{check: updater.CheckResult{
		Current:         "v1.26.1",
		Latest:          "v1.26.2",
		UpdateAvailable: true,
		CanApply:        true,
		ReleaseURL:      "https://example.test/release",
		PublishedAt:     time.Date(2026, 5, 17, 1, 2, 3, 0, time.UTC),
	}}
	controller := NewHTTPController(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	controller.SetUpdateService(stub)

	req := httptest.NewRequest(http.MethodGet, "/update/check?force=1", nil)
	resp := httptest.NewRecorder()
	controller.GetUpdateCheck(NewChiContext(resp, req))

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}
	if !stub.checkForce {
		t.Fatal("expected force=true")
	}
	var body updater.CheckResult
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Latest != "v1.26.2" || !body.CanApply {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestGetUpdateJobReturnsStatus(t *testing.T) {
	stub := &updateServiceStub{status: updater.JobStatus{
		Phase:   updater.PhaseDownloading,
		Current: "v1.26.1",
		Target:  "v1.26.2",
		Message: "Downloading",
	}}
	controller := NewHTTPController(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	controller.SetUpdateService(stub)

	req := httptest.NewRequest(http.MethodGet, "/update/job", nil)
	resp := httptest.NewRecorder()
	controller.GetUpdateJob(NewChiContext(resp, req))

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}
	var body updater.JobStatus
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Phase != updater.PhaseDownloading || body.Target != "v1.26.2" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestApplyUpdateStartsJob(t *testing.T) {
	stub := &updateServiceStub{apply: updater.JobStatus{
		Phase:   updater.PhaseChecking,
		Current: "v1.26.1",
		Target:  "v1.26.2",
		Message: "Update helper started",
	}}
	controller := NewHTTPController(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	controller.SetUpdateService(stub)
	body := bytes.NewBufferString(`{"targetVersion":"v1.26.2"}`)

	req := httptest.NewRequest(http.MethodPost, "/update/apply", body)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	controller.ApplyUpdate(NewChiContext(resp, req))

	if resp.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}
	if stub.applyReq.TargetVersion != "v1.26.2" {
		t.Fatalf("unexpected apply request: %+v", stub.applyReq)
	}
}
