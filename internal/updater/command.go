package updater

import (
	"context"
	"flag"
	"fmt"
	"io"
)

type CommandService interface {
	Check(ctx context.Context, force bool) (CheckResult, error)
	Status(ctx context.Context) (JobStatus, error)
	Apply(ctx context.Context, req ApplyRequest) (JobStatus, error)
}

func RunCommand(ctx context.Context, args []string, stdout io.Writer, service CommandService) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	check := fs.Bool("check", false, "check for updates")
	yes := fs.Bool("yes", false, "apply update without prompting")
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
	if !*yes {
		return fmt.Errorf("refusing to update without --yes")
	}
	status, err := service.Apply(ctx, ApplyRequest{TargetVersion: *version})
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "Update started: %s\n", status.Target)
	_, _ = fmt.Fprintf(stdout, "Status: %s\n", status.Phase)
	return nil
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
