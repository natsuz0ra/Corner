package updater

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type CommandService interface {
	Check(ctx context.Context, force bool) (CheckResult, error)
	Status(ctx context.Context) (JobStatus, error)
	Apply(ctx context.Context, req ApplyRequest) (JobStatus, error)
}

var commandStatusPollInterval = 500 * time.Millisecond

func RunCommand(ctx context.Context, args []string, stdout io.Writer, service CommandService) error {
	return RunCommandWithIO(ctx, args, os.Stdin, stdout, service)
}

func RunCommandWithIO(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer, service CommandService) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	check := fs.Bool("check", false, "check for updates")
	yes := fs.Bool("yes", false, "skip update confirmation")
	version := fs.String("version", "", "target version")
	helper := fs.Bool("helper", false, "run hidden update helper")
	repo := fs.String("repo", "", "release repository")
	statusPath := fs.String("status-file", "", "status file")
	parentPID := fs.Int("parent-pid", 0, "parent pid")
	current := fs.String("current", "", "current version")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *helper {
		return RunHelper(ctx, HelperOptions{
			Repo:          *repo,
			TargetVersion: *version,
			StatusPath:    *statusPath,
			ParentPID:     *parentPID,
			Current:       *current,
		})
	}
	if service == nil {
		return fmt.Errorf("update service is not configured")
	}
	if *check {
		result, err := service.Check(ctx, true)
		if err != nil {
			return err
		}
		printCheckResult(stdout, result)
		return nil
	}
	result, err := service.Check(ctx, true)
	if err != nil {
		return err
	}
	printCheckResult(stdout, result)
	target := strings.TrimSpace(*version)
	if target == "" {
		target = result.Latest
	}
	if !result.UpdateAvailable && strings.TrimSpace(*version) == "" {
		return nil
	}
	if !result.CanApply && strings.TrimSpace(*version) == "" {
		return nil
	}
	if !*yes && !confirmUpdate(stdin, stdout) {
		_, _ = fmt.Fprintln(stdout, "Update cancelled.")
		return nil
	}
	status, err := service.Apply(ctx, ApplyRequest{TargetVersion: target})
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "Update started: %s\n", status.Target)
	printer := newCommandStatusPrinter(stdout)
	printer.Print(status)
	return waitForCommandUpdate(ctx, printer, service)
}

func printCheckResult(w io.Writer, result CheckResult) {
	if result.UpdateAvailable {
		_, _ = fmt.Fprintf(w, "Update available: %s -> %s\n", result.Current, result.Latest)
	} else {
		_, _ = fmt.Fprintf(w, "SlimeBot is up to date: %s\n", result.Current)
	}
	if result.ReleaseURL != "" {
		_, _ = fmt.Fprintf(w, "Release: %s\n", result.ReleaseURL)
	}
	if result.Reason != "" {
		_, _ = fmt.Fprintf(w, "Note: %s\n", result.Reason)
	}
	if result.ManualHint != "" && !result.CanApply {
		_, _ = fmt.Fprintf(w, "Manual: %s\n", result.ManualHint)
	}
	if !result.PublishedAt.IsZero() {
		_, _ = fmt.Fprintf(w, "Published: %s\n", result.PublishedAt.Format("2006-01-02 15:04:05 UTC"))
	}
	if result.AssetName != "" {
		_, _ = fmt.Fprintf(w, "Asset: %s\n", result.AssetName)
	}
}

func confirmUpdate(stdin io.Reader, stdout io.Writer) bool {
	_, _ = fmt.Fprint(stdout, "是否更新? [y/N] ")
	scanner := bufio.NewScanner(stdin)
	if !scanner.Scan() {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes"
}

func waitForCommandUpdate(ctx context.Context, printer *commandStatusPrinter, service CommandService) error {
	ticker := time.NewTicker(commandStatusPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			status, err := service.Status(ctx)
			if err != nil {
				return err
			}
			printer.Print(status)
			if status.Phase == PhaseSucceeded {
				return nil
			}
			if status.Phase == PhaseFailed {
				if status.Error != "" {
					return fmt.Errorf("%s", status.Error)
				}
				return fmt.Errorf("update failed")
			}
			if status.Phase == PhaseIdle {
				return nil
			}
		}
	}
}

func printJobStatus(stdout io.Writer, status JobStatus) {
	newCommandStatusPrinter(stdout).Print(status)
}

type commandStatusPrinter struct {
	stdout     io.Writer
	lastKey    string
	activeLine bool
}

func newCommandStatusPrinter(stdout io.Writer) *commandStatusPrinter {
	return &commandStatusPrinter{stdout: stdout}
}

func (p *commandStatusPrinter) Print(status JobStatus) {
	if status.Phase == "" || status.Phase == PhaseIdle {
		return
	}
	key := commandStatusKey(status)
	if key == p.lastKey {
		return
	}
	p.lastKey = key
	message := status.Message
	if message == "" {
		message = string(status.Phase)
	}
	if status.Phase == PhaseDownloading {
		_, _ = fmt.Fprintf(p.stdout, "\r%s %s", message, formatCommandProgress(status))
		p.activeLine = true
		return
	}
	if p.activeLine {
		_, _ = fmt.Fprint(p.stdout, "\n")
		p.activeLine = false
	}
	if status.Error != "" {
		_, _ = fmt.Fprintf(p.stdout, "%s: %s\n", message, status.Error)
		return
	}
	_, _ = fmt.Fprintln(p.stdout, message)
}

func commandStatusKey(status JobStatus) string {
	return fmt.Sprintf("%s|%s|%s|%d|%d|%d", status.Phase, status.Message, status.Error, status.DownloadedBytes, status.TotalBytes, status.ProgressPercent)
}

func formatCommandProgress(status JobStatus) string {
	if status.TotalBytes > 0 {
		return fmt.Sprintf("[%s] %d%% (%s/%s)", progressBar(status.ProgressPercent, 20), status.ProgressPercent, formatBytes(status.DownloadedBytes), formatBytes(status.TotalBytes))
	}
	if status.DownloadedBytes > 0 {
		return fmt.Sprintf("(%s downloaded)", formatBytes(status.DownloadedBytes))
	}
	return ""
}

func progressBar(percent int, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := percent * width / 100
	return strings.Repeat("#", filled) + strings.Repeat("-", width-filled)
}

func formatBytes(value int64) string {
	if value < 1024 {
		return fmt.Sprintf("%d B", value)
	}
	if value < 1024*1024 {
		return fmt.Sprintf("%.1f KiB", float64(value)/1024)
	}
	return fmt.Sprintf("%.1f MiB", float64(value)/(1024*1024))
}
