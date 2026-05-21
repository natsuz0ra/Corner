package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	memorysvc "slimebot/internal/services/memory"
	"testing"

	"github.com/go-chi/chi/v5"
)

type memoryServiceStub struct {
	snapshot     memorysvc.Snapshot
	clearTarget  memorysvc.Target
	removeTarget memorysvc.Target
	removeIndex  int
}

func (m *memoryServiceStub) Snapshot(_ context.Context) (memorysvc.Snapshot, error) {
	return m.snapshot, nil
}

func (m *memoryServiceStub) Clear(_ context.Context, target memorysvc.Target) (memorysvc.TargetState, error) {
	m.clearTarget = target
	return memorysvc.TargetState{Target: target}, nil
}

func (m *memoryServiceStub) RemoveIndex(_ context.Context, target memorysvc.Target, index int) (memorysvc.TargetState, error) {
	m.removeTarget = target
	m.removeIndex = index
	return memorysvc.TargetState{Target: target}, nil
}

func TestGetMemoryReturnsSnapshot(t *testing.T) {
	stub := &memoryServiceStub{snapshot: memorysvc.Snapshot{
		MemoryEnabled: true,
		Memory: memorysvc.TargetState{
			Target:     memorysvc.TargetMemory,
			Entries:    []string{"项目约定"},
			UsageChars: 4,
			CharLimit:  2200,
			EntryCount: 1,
			Enabled:    true,
		},
	}}
	controller := NewHTTPController(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	controller.SetMemoryService(stub)

	req := httptest.NewRequest(http.MethodGet, "/memory", nil)
	resp := httptest.NewRecorder()
	controller.GetMemory(NewChiContext(resp, req))

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}
	var body memorysvc.Snapshot
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Memory.Entries) != 1 || body.Memory.Entries[0] != "项目约定" {
		t.Fatalf("unexpected snapshot: %#v", body)
	}
}

func TestDeleteMemoryTargetAndEntry(t *testing.T) {
	stub := &memoryServiceStub{}
	controller := NewHTTPController(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	controller.SetMemoryService(stub)

	req := httptest.NewRequest(http.MethodDelete, "/memory/user", nil)
	resp := httptest.NewRecorder()
	controller.ClearMemory(memoryRequestContextWithRecorder(resp, req, map[string]string{"target": "user"}))
	if resp.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}
	if stub.clearTarget != memorysvc.TargetUser {
		t.Fatalf("clear target = %q", stub.clearTarget)
	}

	req = httptest.NewRequest(http.MethodDelete, "/memory/memory/entries/2", nil)
	resp = httptest.NewRecorder()
	controller.DeleteMemoryEntry(memoryRequestContextWithRecorder(resp, req, map[string]string{"target": "memory", "index": "2"}))
	if resp.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}
	if stub.removeTarget != memorysvc.TargetMemory || stub.removeIndex != 2 {
		t.Fatalf("remove = %q/%d", stub.removeTarget, stub.removeIndex)
	}
}

func memoryRequestContextWithRecorder(resp *httptest.ResponseRecorder, req *http.Request, params map[string]string) WebContext {
	routeCtx := chi.NewRouteContext()
	for key, value := range params {
		routeCtx.URLParams.Add(key, value)
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	return NewChiContext(resp, req)
}
