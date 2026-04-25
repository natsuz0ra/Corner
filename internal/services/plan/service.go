package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"slimebot/internal/constants"
	"slimebot/internal/domain"
	"slimebot/internal/runtime"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// PlanFrontmatter is the YAML header written at the top of each plan file.
type PlanFrontmatter struct {
	ID         string `yaml:"id"`
	Title      string `yaml:"title"`
	SessionID  string `yaml:"session_id"`
	Status     string `yaml:"status"`
	CreatedAt  string `yaml:"created_at"`
	ApprovedAt string `yaml:"approved_at,omitempty"`
}

// PlanService manages plan files on disk under ~/.slimebot/plans/.
type PlanService struct {
	plansDir string
}

// NewPlanService creates the service and ensures the plans directory exists.
func NewPlanService() (*PlanService, error) {
	dir := filepath.Join(runtime.SlimeBotHomeDir(), "plans")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create plans dir: %w", err)
	}
	return &PlanService{plansDir: dir}, nil
}

// SavePlan writes a new plan file and returns the Plan.
func (s *PlanService) SavePlan(sessionID, title, content string) (*domain.Plan, error) {
	now := time.Now()
	id := uuid.New().String()
	fm := PlanFrontmatter{
		ID:        id,
		Title:     title,
		SessionID: sessionID,
		Status:    constants.PlanStatusPending,
		CreatedAt: now.Format(time.RFC3339),
	}

	filePath := s.planFilePath(id)
	if err := writePlanFile(filePath, &fm, content); err != nil {
		return nil, fmt.Errorf("save plan: %w", err)
	}

	return &domain.Plan{
		ID:        id,
		Title:     title,
		SessionID: sessionID,
		Status:    constants.PlanStatusPending,
		CreatedAt: fm.CreatedAt,
		Content:   content,
	}, nil
}

// GetPlan reads a single plan by ID.
func (s *PlanService) GetPlan(planID string) (*domain.Plan, error) {
	return readPlanFile(s.planFilePath(planID))
}

// ListPlans returns all plans sorted by creation time (newest first).
func (s *PlanService) ListPlans() ([]domain.Plan, error) {
	entries, err := os.ReadDir(s.plansDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read plans dir: %w", err)
	}

	var plans []domain.Plan
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		p, err := readPlanFile(filepath.Join(s.plansDir, e.Name()))
		if err != nil {
			continue
		}
		plans = append(plans, *p)
	}

	sort.Slice(plans, func(i, j int) bool {
		return plans[i].CreatedAt > plans[j].CreatedAt
	})
	return plans, nil
}

// UpdatePlanStatus changes the plan's status and optionally sets ApprovedAt.
func (s *PlanService) UpdatePlanStatus(planID, status string) (*domain.Plan, error) {
	filePath := s.planFilePath(planID)
	p, err := readPlanFile(filePath)
	if err != nil {
		return nil, err
	}

	p.Status = status
	if status == constants.PlanStatusApproved {
		p.ApprovedAt = time.Now().Format(time.RFC3339)
	}

	fm := PlanFrontmatter{
		ID:         p.ID,
		Title:      p.Title,
		SessionID:  p.SessionID,
		Status:     p.Status,
		CreatedAt:  p.CreatedAt,
		ApprovedAt: p.ApprovedAt,
	}
	if err := writePlanFile(filePath, &fm, p.Content); err != nil {
		return nil, fmt.Errorf("update plan status: %w", err)
	}
	return p, nil
}

// DeletePlan removes a plan file.
func (s *PlanService) DeletePlan(planID string) error {
	filePath := s.planFilePath(planID)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete plan: %w", err)
	}
	return nil
}

// GetPlansBySession returns plans filtered by session ID.
func (s *PlanService) GetPlansBySession(sessionID string) ([]domain.Plan, error) {
	all, err := s.ListPlans()
	if err != nil {
		return nil, err
	}
	var filtered []domain.Plan
	for _, p := range all {
		if p.SessionID == sessionID {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

func (s *PlanService) planFilePath(planID string) string {
	return filepath.Join(s.plansDir, planID+".md")
}

func writePlanFile(filePath string, fm *PlanFrontmatter, content string) error {
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return fmt.Errorf("marshal frontmatter: %w", err)
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.Write(yamlBytes)
	b.WriteString("---\n\n")
	b.WriteString(content)

	return os.WriteFile(filePath, []byte(b.String()), 0o644)
}

func readPlanFile(filePath string) (*domain.Plan, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read plan file: %w", err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("invalid plan file format: missing frontmatter")
	}

	end := strings.Index(content[4:], "\n---\n")
	if end < 0 {
		return nil, fmt.Errorf("invalid plan file format: unclosed frontmatter")
	}

	fmStr := content[4 : end+4]
	body := strings.TrimSpace(content[end+8:])

	var fm PlanFrontmatter
	if err := yaml.Unmarshal([]byte(fmStr), &fm); err != nil {
		return nil, fmt.Errorf("parse plan frontmatter: %w", err)
	}

	return &domain.Plan{
		ID:         fm.ID,
		Title:      fm.Title,
		SessionID:  fm.SessionID,
		Status:     fm.Status,
		CreatedAt:  fm.CreatedAt,
		ApprovedAt: fm.ApprovedAt,
		Content:    body,
	}, nil
}
