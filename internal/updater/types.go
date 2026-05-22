package updater

import "time"

const DefaultRepo = "natsuz0ra/SlimeBot"

type Phase string

const (
	PhaseIdle        Phase = "idle"
	PhaseChecking    Phase = "checking"
	PhaseDownloading Phase = "downloading"
	PhaseInstalling  Phase = "installing"
	PhaseRestarting  Phase = "restarting"
	PhaseSucceeded   Phase = "succeeded"
	PhaseFailed      Phase = "failed"
)

type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"downloadUrl"`
}

type Release struct {
	TagName     string    `json:"tagName"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"htmlUrl"`
	PublishedAt time.Time `json:"publishedAt"`
	Prerelease  bool      `json:"prerelease"`
	Assets      []Asset   `json:"assets"`
}

type CheckResult struct {
	Current         string    `json:"current"`
	Latest          string    `json:"latest"`
	UpdateAvailable bool      `json:"updateAvailable"`
	CanApply        bool      `json:"canApply"`
	Reason          string    `json:"reason"`
	ReleaseName     string    `json:"releaseName"`
	ReleaseNotes    string    `json:"releaseNotes"`
	ReleaseURL      string    `json:"releaseUrl"`
	PublishedAt     time.Time `json:"publishedAt"`
	AssetName       string    `json:"assetName"`
	ManualHint      string    `json:"manualHint"`
}

type JobStatus struct {
	Phase           Phase     `json:"phase"`
	Current         string    `json:"current"`
	Target          string    `json:"target"`
	Message         string    `json:"message"`
	Error           string    `json:"error,omitempty"`
	ManualHint      string    `json:"manualHint,omitempty"`
	DownloadedBytes int64     `json:"downloadedBytes,omitempty"`
	TotalBytes      int64     `json:"totalBytes,omitempty"`
	ProgressPercent int       `json:"progressPercent,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type ApplyRequest struct {
	TargetVersion string `json:"targetVersion"`
}

type HelperOptions struct {
	Repo          string
	TargetVersion string
	StatusPath    string
	ParentPID     int
	Current       string
}
