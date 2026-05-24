package updater

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

type commandServiceStub struct {
	checkCalled bool
	applyReq    ApplyRequest
	check       CheckResult
	apply       JobStatus
	statuses    []JobStatus
}

func (s *commandServiceStub) Check(_ context.Context, force bool) (CheckResult, error) {
	s.checkCalled = force
	return s.check, nil
}

func (s *commandServiceStub) Status(_ context.Context) (JobStatus, error) {
	if len(s.statuses) > 0 {
		status := s.statuses[0]
		s.statuses = s.statuses[1:]
		return status, nil
	}
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
	stub := &commandServiceStub{
		check: CheckResult{
			Current:         "v1.26.1",
			Latest:          "v1.26.2",
			UpdateAvailable: true,
			CanApply:        true,
		},
		apply: JobStatus{
			Phase:  PhaseChecking,
			Target: "v1.26.2",
		},
		statuses: []JobStatus{{Phase: PhaseSucceeded, Target: "v1.26.2", Message: "Update installed successfully"}},
	}
	var stdout bytes.Buffer

	oldPoll := commandStatusPollInterval
	commandStatusPollInterval = 1
	defer func() { commandStatusPollInterval = oldPoll }()

	if err := RunCommandWithIO(context.Background(), []string{"--version", "v1.26.2", "--yes"}, strings.NewReader(""), &stdout, stub); err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}
	if stub.applyReq.TargetVersion != "v1.26.2" {
		t.Fatalf("unexpected apply request: %+v", stub.applyReq)
	}
	if !strings.Contains(stdout.String(), "Update started") {
		t.Fatalf("unexpected output: %q", stdout.String())
	}
}

func TestRunCommandPromptsBeforeApplyingUpdate(t *testing.T) {
	stub := &commandServiceStub{check: CheckResult{
		Current:         "v1.26.1",
		Latest:          "v1.26.2",
		UpdateAvailable: true,
		CanApply:        true,
		ReleaseURL:      "https://example.test/release",
		AssetName:       "slimebot-v1.26.2-darwin-arm64.tar.gz",
	}}
	var stdout bytes.Buffer

	if err := RunCommandWithIO(context.Background(), nil, strings.NewReader("n\n"), &stdout, stub); err != nil {
		t.Fatalf("RunCommandWithIO failed: %v", err)
	}
	if stub.applyReq.TargetVersion != "" {
		t.Fatalf("apply should not be called, got request: %+v", stub.applyReq)
	}
	out := stdout.String()
	if !strings.Contains(out, "Update available: v1.26.1 -> v1.26.2") || !strings.Contains(out, "是否更新? [y/N]") || !strings.Contains(out, "Update cancelled") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRunCommandYesSkipsPromptAndWaitsForCompletion(t *testing.T) {
	stub := &commandServiceStub{
		check: CheckResult{
			Current:         "v1.26.1",
			Latest:          "v1.26.2",
			UpdateAvailable: true,
			CanApply:        true,
		},
		apply: JobStatus{Phase: PhaseChecking, Target: "v1.26.2"},
		statuses: []JobStatus{
			{Phase: PhaseDownloading, Target: "v1.26.2", Message: "Downloading", DownloadedBytes: 5, TotalBytes: 10, ProgressPercent: 50},
			{Phase: PhaseSucceeded, Target: "v1.26.2", Message: "Update installed successfully"},
		},
	}
	var stdout bytes.Buffer

	oldPoll := commandStatusPollInterval
	commandStatusPollInterval = 1
	defer func() { commandStatusPollInterval = oldPoll }()

	if err := RunCommandWithIO(context.Background(), []string{"--yes"}, io.NopCloser(strings.NewReader("")), &stdout, stub); err != nil {
		t.Fatalf("RunCommandWithIO failed: %v", err)
	}
	if stub.applyReq.TargetVersion != "v1.26.2" {
		t.Fatalf("unexpected apply request: %+v", stub.applyReq)
	}
	out := stdout.String()
	if !strings.Contains(out, "Downloading") || !strings.Contains(out, "50%") || !strings.Contains(out, "Update installed successfully") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestPrintJobStatusesRefreshesDownloadLineAndSuppressesDuplicates(t *testing.T) {
	var stdout bytes.Buffer
	printer := newCommandStatusPrinter(&stdout)

	printer.Print(JobStatus{Phase: PhaseDownloading, Message: "Downloading", DownloadedBytes: 1, TotalBytes: 10, ProgressPercent: 10})
	printer.Print(JobStatus{Phase: PhaseDownloading, Message: "Downloading", DownloadedBytes: 5, TotalBytes: 10, ProgressPercent: 50})
	printer.Print(JobStatus{Phase: PhaseDownloading, Message: "Downloading", DownloadedBytes: 10, TotalBytes: 10, ProgressPercent: 100})
	printer.Print(JobStatus{Phase: PhaseInstalling, Message: "Installing"})
	printer.Print(JobStatus{Phase: PhaseInstalling, Message: "Installing"})
	printer.Print(JobStatus{Phase: PhaseRestarting, Message: "Restarting"})
	printer.Print(JobStatus{Phase: PhaseRestarting, Message: "Restarting"})
	printer.Print(JobStatus{Phase: PhaseSucceeded, Message: "Done"})

	out := stdout.String()
	if strings.Count(out, "\rDownloading") != 3 || !strings.Contains(out, "100%") {
		t.Fatalf("download progress should refresh in place, got %q", out)
	}
	if strings.Count(out, "Installing\n") != 1 || strings.Count(out, "Restarting\n") != 1 {
		t.Fatalf("duplicate terminal statuses should be suppressed, got %q", out)
	}
	if !strings.Contains(out, "\nInstalling\n") || !strings.Contains(out, "Done\n") {
		t.Fatalf("phase changes should end on readable lines, got %q", out)
	}
}
