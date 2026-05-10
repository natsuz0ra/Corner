package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"slimebot/internal/domain"
)

func TestFileSystemSkillStore_ListSkillsReadsDirectory(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "alpha", "Alpha skill")
	writeSkill(t, root, "beta", "Beta skill")

	store := NewFileSystemSkillStore(root)
	items, err := store.ListSkills()
	if err != nil {
		t.Fatalf("ListSkills failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(items))
	}
	if items[0].ID == "" || items[0].RelativePath == "" {
		t.Fatalf("expected stable metadata, got %#v", items[0])
	}
}

func TestFileSystemSkillStore_GetSkillByName(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "alpha", "Alpha skill")

	store := NewFileSystemSkillStore(root)
	item, err := store.GetSkillByName("alpha")
	if err != nil {
		t.Fatalf("GetSkillByName failed: %v", err)
	}
	if item == nil || item.Name != "alpha" || item.Description != "Alpha skill" {
		t.Fatalf("unexpected skill: %#v", item)
	}
}

func TestFileSystemSkillStore_DeleteSkillRemovesDirectory(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "alpha", "Alpha skill")

	store := NewFileSystemSkillStore(root)
	if err := store.DeleteSkill("alpha"); err != nil {
		t.Fatalf("DeleteSkill failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "alpha")); !os.IsNotExist(err) {
		t.Fatalf("expected directory deleted, stat err=%v", err)
	}
}

func TestSkillRuntimeService_BuildCatalogPromptUsesDirectoryStore(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "beta", "Beta skill")
	writeSkill(t, root, "alpha", "Alpha skill")

	store := NewFileSystemSkillStore(root)
	svc := NewSkillRuntimeService(store, root)
	prompt, skills, err := svc.BuildCatalogPrompt()
	if err != nil {
		t.Fatalf("BuildCatalogPrompt failed: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
	if !strings.Contains(prompt, "alpha") || !strings.Contains(prompt, "Alpha skill") {
		t.Fatalf("unexpected prompt: %s", prompt)
	}
	if skills[0].Name != "alpha" || skills[1].Name != "beta" {
		t.Fatalf("expected sorted skills [alpha beta], got [%s %s]", skills[0].Name, skills[1].Name)
	}
	if strings.Index(prompt, "<name>alpha</name>") > strings.Index(prompt, "<name>beta</name>") {
		t.Fatalf("expected alpha before beta in prompt: %s", prompt)
	}
}

func TestSkillRuntimeService_DiscoversExternalGlobalSkills(t *testing.T) {
	root := t.TempDir()
	stateDir := t.TempDir()
	codexRoot := t.TempDir()
	claudeRoot := t.TempDir()
	hermesRoot := t.TempDir()
	writeSkill(t, filepath.Join(codexRoot, "skills"), "shared", "Codex skill")
	writeSkill(t, filepath.Join(claudeRoot, "skills"), "shared", "Claude skill")
	writeSkill(t, filepath.Join(hermesRoot, "skills"), "planning", "Hermes skill")

	store := NewFileSystemSkillStore(root)
	svc := NewSkillRuntimeServiceWithOptions(store, root, SkillRuntimeOptions{
		StateDir: stateDir,
		Sources: []SkillSource{
			{Provider: "codex", Label: "Codex", Scope: "global", Root: filepath.Join(codexRoot, "skills"), ReadOnly: true},
			{Provider: "claude", Label: "Claude Code", Scope: "global", Root: filepath.Join(claudeRoot, "skills"), ReadOnly: true},
			{Provider: "hermes", Label: "Hermes Agent", Scope: "global", Root: filepath.Join(hermesRoot, "skills"), ReadOnly: true},
		},
	})

	prompt, skills, err := svc.BuildCatalogPrompt()
	if err != nil {
		t.Fatalf("BuildCatalogPrompt failed: %v", err)
	}
	if len(skills) != 3 {
		t.Fatalf("expected 3 skills, got %d: %#v", len(skills), skills)
	}
	assertSkillID(t, skills, "codex:global:shared")
	assertSkillID(t, skills, "claude:global:shared")
	assertSkillID(t, skills, "hermes:global:planning")
	if !strings.Contains(prompt, "codex:global:shared") || !strings.Contains(prompt, "Claude skill") {
		t.Fatalf("expected external skills in catalog, got: %s", prompt)
	}
}

func TestSkillRuntimeService_DisabledExternalSkillIsHiddenAndCannotActivate(t *testing.T) {
	root := t.TempDir()
	stateDir := t.TempDir()
	codexRoot := t.TempDir()
	writeSkill(t, filepath.Join(codexRoot, "skills"), "alpha", "Codex alpha")

	store := NewFileSystemSkillStore(root)
	svc := NewSkillRuntimeServiceWithOptions(store, root, SkillRuntimeOptions{
		StateDir: stateDir,
		Sources: []SkillSource{
			{Provider: "codex", Label: "Codex", Scope: "global", Root: filepath.Join(codexRoot, "skills"), ReadOnly: true},
		},
	})
	if err := svc.SetSkillEnabled("codex:global:alpha", false); err != nil {
		t.Fatalf("SetSkillEnabled failed: %v", err)
	}

	prompt, skills, err := svc.BuildCatalogPrompt()
	if err != nil {
		t.Fatalf("BuildCatalogPrompt failed: %v", err)
	}
	if len(skills) != 0 || strings.Contains(prompt, "alpha") {
		t.Fatalf("disabled skill should be hidden, skills=%#v prompt=%q", skills, prompt)
	}
	if _, _, err := svc.ActivateSkill("codex:global:alpha", map[string]struct{}{}); err == nil {
		t.Fatal("expected disabled skill activation to fail")
	}
}

func TestSkillRuntimeService_ProjectSourcesOnlyWhenConfigured(t *testing.T) {
	root := t.TempDir()
	stateDir := t.TempDir()
	project := t.TempDir()
	writeSkill(t, filepath.Join(project, ".codex", "skills"), "project-skill", "Project skill")

	webSvc := NewSkillRuntimeServiceWithOptions(NewFileSystemSkillStore(root), root, SkillRuntimeOptions{
		StateDir: stateDir,
		Sources:  []SkillSource{},
	})
	if _, skills, err := webSvc.BuildCatalogPrompt(); err != nil || len(skills) != 0 {
		t.Fatalf("web-style service should not discover project skills, skills=%#v err=%v", skills, err)
	}

	cliSvc := NewSkillRuntimeServiceWithOptions(NewFileSystemSkillStore(root), root, SkillRuntimeOptions{
		StateDir: stateDir,
		Sources: []SkillSource{
			{Provider: "codex", Label: "Codex", Scope: "project", Root: filepath.Join(project, ".codex", "skills"), ReadOnly: true},
		},
	})
	if _, skills, err := cliSvc.BuildCatalogPrompt(); err != nil || len(skills) != 1 || skills[0].ID != "codex:project:project-skill" {
		t.Fatalf("cli-style service should discover project skill, skills=%#v err=%v", skills, err)
	}
}

func TestSkillRuntimeService_ExternalSkillsAreReadOnly(t *testing.T) {
	root := t.TempDir()
	stateDir := t.TempDir()
	codexRoot := t.TempDir()
	writeSkill(t, filepath.Join(codexRoot, "skills"), "alpha", "Codex alpha")
	writeSkill(t, root, "local", "Local skill")

	svc := NewSkillRuntimeServiceWithOptions(NewFileSystemSkillStore(root), root, SkillRuntimeOptions{
		StateDir: stateDir,
		Sources: []SkillSource{
			{Provider: "codex", Label: "Codex", Scope: "global", Root: filepath.Join(codexRoot, "skills"), ReadOnly: true},
		},
	})
	if err := svc.DeleteSkillByID("codex:global:alpha"); err == nil {
		t.Fatal("expected external skill delete to fail")
	}
	if err := svc.DeleteSkillByID("local"); err != nil {
		t.Fatalf("expected local skill delete to succeed: %v", err)
	}
}

func assertSkillID(t *testing.T, skills []domain.Skill, id string) {
	t.Helper()
	for _, item := range skills {
		if item.ID == id {
			return
		}
	}
	t.Fatalf("expected skill id %s in %#v", id, skills)
}

func writeSkill(t *testing.T, root, name, desc string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := "---\nname: " + name + "\ndescription: " + desc + "\n---\n\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write skill failed: %v", err)
	}
}
