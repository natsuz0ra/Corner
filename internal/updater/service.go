package updater

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	sbruntime "slimebot/internal/runtime"
)

type ServiceOptions struct {
	CurrentVersion string
	Repo           string
	APIBaseURL     string
	HTTPClient     *http.Client
	GOOS           string
	GOARCH         string
	StatusPath     string
	HelperStarter  func(ctx context.Context, opts HelperOptions) error
}

type Service struct {
	currentVersion string
	repo           string
	apiBaseURL     string
	httpClient     *http.Client
	goos           string
	goarch         string
	status         *StatusStore
	helperStarter  func(ctx context.Context, opts HelperOptions) error
}

func NewService(opts ServiceOptions) *Service {
	repo := strings.TrimSpace(opts.Repo)
	if repo == "" {
		repo = strings.TrimSpace(os.Getenv("SLIMEBOT_REPO"))
	}
	if repo == "" {
		repo = DefaultRepo
	}
	apiBaseURL := strings.TrimRight(strings.TrimSpace(opts.APIBaseURL), "/")
	if apiBaseURL == "" {
		apiBaseURL = "https://api.github.com"
	}
	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	goos := strings.TrimSpace(opts.GOOS)
	if goos == "" {
		goos = runtime.GOOS
	}
	goarch := strings.TrimSpace(opts.GOARCH)
	if goarch == "" {
		goarch = runtime.GOARCH
	}
	statusPath := strings.TrimSpace(opts.StatusPath)
	if statusPath == "" {
		statusPath = DefaultStatusPath()
	}
	helperStarter := opts.HelperStarter
	if helperStarter == nil {
		helperStarter = StartHelperProcess
	}
	return &Service{
		currentVersion: strings.TrimSpace(opts.CurrentVersion),
		repo:           repo,
		apiBaseURL:     apiBaseURL,
		httpClient:     client,
		goos:           goos,
		goarch:         goarch,
		status:         NewStatusStore(statusPath),
		helperStarter:  helperStarter,
	}
}

func DefaultStatusPath() string {
	return filepath.Join(sbruntime.SlimeBotHomeDir(), "storage", "update-status.json")
}

func ManualUpdateHint(version string) string {
	trimmed := strings.TrimSpace(version)
	if trimmed == "" {
		trimmed = "latest"
	}
	return fmt.Sprintf("slimebot update --version %s", trimmed)
}

func (s *Service) Check(ctx context.Context, _ bool) (CheckResult, error) {
	if fakeUpdateEnabled() {
		return s.fakeCheck(), nil
	}
	release, err := s.fetchLatestRelease(ctx)
	if err != nil {
		return CheckResult{}, err
	}
	current := strings.TrimSpace(s.currentVersion)
	if current == "" {
		current = "dev"
	}
	assetName := AssetNameForPlatform(release.TagName, s.goos, s.goarch)
	_, hasAsset := FindAsset(release, assetName)
	cmp := CompareVersions(release.TagName, current)
	_, currentKnown := parseSemver(current)
	updateAvailable := cmp > 0
	if !currentKnown {
		updateAvailable = true
	}
	result := CheckResult{
		Current:         current,
		Latest:          release.TagName,
		UpdateAvailable: updateAvailable,
		CanApply:        updateAvailable && currentKnown && hasAsset,
		ReleaseName:     release.Name,
		ReleaseNotes:    release.Body,
		ReleaseURL:      release.HTMLURL,
		PublishedAt:     release.PublishedAt,
		AssetName:       assetName,
		ManualHint:      ManualUpdateHint(release.TagName),
	}
	if !updateAvailable {
		result.Reason = "already latest"
		return result, nil
	}
	if !currentKnown {
		result.Reason = fmt.Sprintf("current version %q cannot be updated automatically", current)
		return result, nil
	}
	if !hasAsset {
		result.Reason = fmt.Sprintf("release asset %q was not found", assetName)
		return result, nil
	}
	return result, nil
}

func (s *Service) Status(ctx context.Context) (JobStatus, error) {
	return s.status.Read(ctx)
}

func (s *Service) Apply(ctx context.Context, req ApplyRequest) (JobStatus, error) {
	if fakeUpdateEnabled() {
		return s.fakeApply(ctx, req)
	}
	target := strings.TrimSpace(req.TargetVersion)
	current := strings.TrimSpace(s.currentVersion)
	if current == "" {
		current = "dev"
	}
	if target == "" {
		if _, ok := parseSemver(current); !ok {
			return JobStatus{}, fmt.Errorf("current version %q cannot be updated automatically; use %s", current, ManualUpdateHint("vX.Y.Z"))
		}
		check, err := s.Check(ctx, true)
		if err != nil {
			return JobStatus{}, err
		}
		if !check.UpdateAvailable {
			return JobStatus{}, fmt.Errorf("SlimeBot is already up to date")
		}
		if !check.CanApply {
			return JobStatus{}, errors.New(check.Reason)
		}
		target = check.Latest
	}

	status := JobStatus{
		Phase:      PhaseChecking,
		Current:    current,
		Target:     target,
		Message:    "Update helper started",
		ManualHint: ManualUpdateHint(target),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := s.status.Write(ctx, status); err != nil {
		return JobStatus{}, err
	}
	if err := s.helperStarter(ctx, HelperOptions{
		Repo:          s.repo,
		TargetVersion: target,
		StatusPath:    s.status.Path(),
		ParentPID:     os.Getpid(),
		Current:       current,
	}); err != nil {
		failed := status
		failed.Phase = PhaseFailed
		failed.Error = err.Error()
		failed.Message = "Failed to start update helper"
		failed.UpdatedAt = time.Now().UTC()
		_ = s.status.Write(context.Background(), failed)
		return JobStatus{}, err
	}
	return status, nil
}

func StartHelperProcess(_ context.Context, opts HelperOptions) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	args := []string{
		"update",
		"--helper",
		"--version", strings.TrimSpace(opts.TargetVersion),
		"--repo", strings.TrimSpace(opts.Repo),
		"--status-file", strings.TrimSpace(opts.StatusPath),
		"--parent-pid", strconv.Itoa(opts.ParentPID),
		"--current", strings.TrimSpace(opts.Current),
	}
	cmd := exec.Command(exe, args...)
	cmd.Env = os.Environ()
	detachProcess(cmd)
	return cmd.Start()
}

func (s *Service) fetchLatestRelease(ctx context.Context) (Release, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/releases/latest", s.apiBaseURL, strings.Trim(s.repo, "/"))
	return s.fetchRelease(ctx, endpoint)
}

func fakeUpdateEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("SLIMEBOT_FAKE_UPDATE")))
	return value == "1" || value == "true" || value == "yes"
}

func (s *Service) fakeCheck() CheckResult {
	current := strings.TrimSpace(s.currentVersion)
	if current == "" || current == "dev" {
		current = "v1.28.0"
	}
	latest := "v9.99.9"
	return CheckResult{
		Current:         current,
		Latest:          latest,
		UpdateAvailable: true,
		CanApply:        true,
		ReleaseName:     "SlimeBot " + latest,
		ReleaseNotes:    "## What's New\n- Improved the update confirmation flow.\n- Added download progress indicators.\n- Added clearer status prompts during installation.",
		ReleaseURL:      "https://github.com/natsuz0ra/SlimeBot/releases/tag/" + latest,
		PublishedAt:     time.Now().UTC(),
		AssetName:       AssetNameForPlatform(latest, s.goos, s.goarch),
		ManualHint:      ManualUpdateHint(latest),
	}
}

func (s *Service) fakeApply(ctx context.Context, req ApplyRequest) (JobStatus, error) {
	current := strings.TrimSpace(s.currentVersion)
	if current == "" || current == "dev" {
		current = "v1.28.0"
	}
	target := strings.TrimSpace(req.TargetVersion)
	if target == "" {
		target = "v9.99.9"
	}
	status := JobStatus{
		Phase:      PhaseChecking,
		Current:    current,
		Target:     target,
		Message:    "Update started",
		ManualHint: ManualUpdateHint(target),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := s.status.Write(ctx, status); err != nil {
		return JobStatus{}, err
	}
	go s.runFakeUpdate(current, target)
	return status, nil
}

func (s *Service) runFakeUpdate(current string, target string) {
	total := int64(32 * 1024 * 1024)
	write := func(status JobStatus) {
		status.Current = current
		status.Target = target
		status.ManualHint = ManualUpdateHint(target)
		status.UpdatedAt = time.Now().UTC()
		_ = s.status.Write(context.Background(), status)
	}
	for percent := 0; percent <= 100; percent += 10 {
		downloaded := total * int64(percent) / 100
		write(JobStatus{
			Phase:           PhaseDownloading,
			Message:         "Downloading update package",
			DownloadedBytes: downloaded,
			TotalBytes:      total,
			ProgressPercent: percent,
		})
		if percent == 100 {
			time.Sleep(downloadCompleteDisplayDelay)
		} else {
			time.Sleep(350 * time.Millisecond)
		}
	}
	write(JobStatus{Phase: PhaseInstalling, Message: "Installing update package"})
	time.Sleep(900 * time.Millisecond)
	write(JobStatus{Phase: PhaseRestarting, Message: "Restarting SlimeBot service"})
	time.Sleep(900 * time.Millisecond)
	write(JobStatus{Phase: PhaseSucceeded, Message: "Update installed successfully"})
}

func (s *Service) fetchReleaseByTag(ctx context.Context, tag string) (Release, error) {
	trimmed := strings.TrimSpace(tag)
	if trimmed == "" {
		return Release{}, fmt.Errorf("release tag is empty")
	}
	endpoint := fmt.Sprintf("%s/repos/%s/releases/tags/%s", s.apiBaseURL, strings.Trim(s.repo, "/"), url.PathEscape(trimmed))
	return s.fetchRelease(ctx, endpoint)
}

func (s *Service) fetchRelease(ctx context.Context, endpoint string) (Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "SlimeBot-Updater")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		return Release{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Release{}, fmt.Errorf("release check failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return ParseLatestRelease(data)
}
