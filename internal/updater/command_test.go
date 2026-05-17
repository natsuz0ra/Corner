package updater

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

type commandServiceStub struct {
	checkCalled bool
	applyReq    ApplyRequest
	check       CheckResult
	apply       JobStatus
}

func (s *commandServiceStub) Check(_ context.Context, force bool) (CheckResult, error) {
	s.checkCalled = force
	return s.check, nil
}

func (s *commandServiceStub) Status(_ context.Context) (JobStatus, error) {
	return JobStatus{Phase: PhaseIdle}, nil
}

func (s *commandServiceStub) Apply(_ context.Context, req ApplyRequest) (JobStatus, error) {
	s.applyReq = req
	return s.apply, nil
}

func TestRunCommandCheckPrintsVersionSummary(t *testing.T) {
	stub := &commandServiceStub{check: CheckResult{
		Current:         "v1.26.1",
		Latest:          "v1.26.2",
		UpdateAvailable: true,
		CanApply:        true,
		ReleaseURL:      "https://example.test/release",
	}}
	var stdout bytes.Buffer

	if err := RunCommand(context.Background(), []string{"--check"}, &stdout, stub); err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}
	if !stub.checkCalled {
		t.Fatal("expected forced check")
	}
	out := stdout.String()
	if !strings.Contains(out, "v1.26.1 -> v1.26.2") || !strings.Contains(out, "https://example.test/release") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRunCommandApplyStartsUpdate(t *testing.T) {
	stub := &commandServiceStub{apply: JobStatus{
		Phase:  PhaseChecking,
		Target: "v1.26.2",
	}}
	var stdout bytes.Buffer

	if err := RunCommand(context.Background(), []string{"--version", "v1.26.2"}, &stdout, stub); err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}
	if stub.applyReq.TargetVersion != "v1.26.2" {
		t.Fatalf("unexpected apply request: %+v", stub.applyReq)
	}
	if !strings.Contains(stdout.String(), "Update started") {
		t.Fatalf("unexpected output: %q", stdout.String())
	}
}
