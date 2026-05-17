package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func RunHelper(ctx context.Context, opts HelperOptions) error {
	statusPath := strings.TrimSpace(opts.StatusPath)
	if statusPath == "" {
		statusPath = DefaultStatusPath()
	}
	store := NewStatusStore(statusPath)
	current := strings.TrimSpace(opts.Current)
	target := strings.TrimSpace(opts.TargetVersion)
	repo := strings.TrimSpace(opts.Repo)
	if repo == "" {
		repo = strings.TrimSpace(os.Getenv("SLIMEBOT_REPO"))
	}
	if repo == "" {
		repo = DefaultRepo
	}
	writeStatus := func(phase Phase, message string, errText string) {
		_ = store.Write(context.Background(), JobStatus{
			Phase:      phase,
			Current:    current,
			Target:     target,
			Message:    message,
			Error:      errText,
			ManualHint: ManualUpdateHint(target),
			UpdatedAt:  time.Now().UTC(),
		})
	}
	fail := func(message string, err error) error {
		errText := ""
		if err != nil {
			errText = err.Error()
		}
		writeStatus(PhaseFailed, message, errText)
		if err == nil {
			return errors.New(message)
		}
		return err
	}

	if target == "" {
		return fail("Target version is required", nil)
	}
	waitForParentExit(opts.ParentPID, 12*time.Second)

	writeStatus(PhaseChecking, "Checking release assets", "")
	service := NewService(ServiceOptions{CurrentVersion: current, Repo: repo, StatusPath: statusPath})
	release, err := service.fetchReleaseByTag(ctx, target)
	if err != nil {
		return fail("Failed to check target release", err)
	}
	assetName := AssetNameForPlatform(target, runtime.GOOS, runtime.GOARCH)
	asset, ok := FindAsset(release, assetName)
	if !ok {
		return fail(fmt.Sprintf("Release asset %q was not found", assetName), nil)
	}

	tmpDir, err := os.MkdirTemp("", "slimebot-update-*")
	if err != nil {
		return fail("Failed to create temporary update directory", err)
	}
	defer os.RemoveAll(tmpDir)

	writeStatus(PhaseDownloading, "Downloading update package", "")
	archivePath := filepath.Join(tmpDir, asset.Name)
	if err := downloadToFile(ctx, service.httpClient, asset.DownloadURL, archivePath); err != nil {
		return fail("Failed to download update package", err)
	}

	writeStatus(PhaseInstalling, "Installing update package", "")
	packageDir, err := extractArchive(archivePath, tmpDir)
	if err != nil {
		return fail("Failed to extract update package", err)
	}
	_ = runServiceAction("stop")
	if err := runInstallScript(ctx, packageDir); err != nil {
		return fail("Failed to run install script", err)
	}

	writeStatus(PhaseRestarting, "Restarting SlimeBot service", "")
	if err := runServiceAction("start"); err != nil {
		writeStatus(PhaseSucceeded, "Update installed. Restart SlimeBot manually if the service is not running.", "")
		return nil
	}
	writeStatus(PhaseSucceeded, "Update installed successfully", "")
	return nil
}

func waitForParentExit(pid int, timeout time.Duration) {
	if pid <= 0 {
		return
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !processExists(pid) {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func processExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return process.Signal(syscall.Signal(0)) == nil
}

func downloadToFile(ctx context.Context, client *http.Client, url string, path string) error {
	if strings.TrimSpace(url) == "" {
		return fmt.Errorf("download url is empty")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func extractArchive(archivePath, destDir string) (string, error) {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath, destDir)
	}
	if strings.HasSuffix(archivePath, ".tar.gz") {
		return extractTarGz(archivePath, destDir)
	}
	return "", fmt.Errorf("unsupported archive format: %s", archivePath)
}

func extractZip(archivePath, destDir string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	var top string
	for _, file := range reader.File {
		target, err := safeJoin(destDir, file.Name)
		if err != nil {
			return "", err
		}
		top = topDir(top, file.Name)
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, file.Mode()); err != nil {
				return "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return "", err
		}
		src, err := file.Open()
		if err != nil {
			return "", err
		}
		if err := writeExtractedFile(target, src, file.Mode()); err != nil {
			_ = src.Close()
			return "", err
		}
		_ = src.Close()
	}
	return filepath.Join(destDir, top), nil
}

func extractTarGz(archivePath, destDir string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gz.Close()
	reader := tar.NewReader(gz)
	var top string
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		target, err := safeJoin(destDir, header.Name)
		if err != nil {
			return "", err
		}
		top = topDir(top, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return "", err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return "", err
			}
			if err := writeExtractedFile(target, reader, os.FileMode(header.Mode)); err != nil {
				return "", err
			}
		}
	}
	return filepath.Join(destDir, top), nil
}

func safeJoin(base, name string) (string, error) {
	target := filepath.Join(base, name)
	cleanBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	cleanTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if cleanTarget != cleanBase && !strings.HasPrefix(cleanTarget, cleanBase+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive entry escapes destination: %s", name)
	}
	return cleanTarget, nil
}

func topDir(current, name string) string {
	clean := filepath.Clean(name)
	if clean == "." || strings.HasPrefix(clean, "..") {
		return current
	}
	part := strings.Split(clean, string(os.PathSeparator))[0]
	if current == "" {
		return part
	}
	if current == part {
		return current
	}
	return current
}

func writeExtractedFile(target string, src io.Reader, mode os.FileMode) error {
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

func runInstallScript(ctx context.Context, packageDir string) error {
	if runtime.GOOS == "windows" {
		script := filepath.Join(packageDir, "install.ps1")
		cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	script := filepath.Join(packageDir, "install.sh")
	cmd := exec.CommandContext(ctx, "sh", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runServiceAction(action string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "service", action)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}
