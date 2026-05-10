package sandbox

import (
	"path/filepath"
	"testing"
)

func TestPolicyWorkspaceWriteAllowsOnlyWritableRoots(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()
	policy, err := NewPolicy(Config{
		Mode:          ModeWorkspaceWrite,
		CWD:           root,
		WritableRoots: []string{filepath.Join(root, "generated")},
		Network:       NetworkPolicy{Enabled: false},
	})
	if err != nil {
		t.Fatalf("new policy: %v", err)
	}

	if err := policy.CheckWrite(filepath.Join(root, "main.go")); err != nil {
		t.Fatalf("cwd write should be allowed: %v", err)
	}
	if err := policy.CheckWrite(filepath.Join(root, "generated", "out.txt")); err != nil {
		t.Fatalf("writable root write should be allowed: %v", err)
	}
	if err := policy.CheckWrite(filepath.Join(other, "escape.txt")); err == nil {
		t.Fatal("workspace-write should reject writes outside cwd/writable roots")
	}
}

func TestPolicyReadOnlyRejectsWrites(t *testing.T) {
	root := t.TempDir()
	policy, err := NewPolicy(Config{Mode: ModeReadOnly, CWD: root})
	if err != nil {
		t.Fatalf("new policy: %v", err)
	}
	if err := policy.CheckRead(filepath.Join(root, "main.go")); err != nil {
		t.Fatalf("read-only should allow reads: %v", err)
	}
	if err := policy.CheckWrite(filepath.Join(root, "main.go")); err == nil {
		t.Fatal("read-only should reject writes")
	}
}

func TestPolicyDenyPathOverridesWritableRoot(t *testing.T) {
	root := t.TempDir()
	secret := filepath.Join(root, ".env")
	policy, err := NewPolicy(Config{
		Mode:      ModeWorkspaceWrite,
		CWD:       root,
		DenyPaths: []string{secret},
	})
	if err != nil {
		t.Fatalf("new policy: %v", err)
	}
	if err := policy.CheckRead(secret); err == nil {
		t.Fatal("deny path should reject reads")
	}
	if err := policy.CheckWrite(secret); err == nil {
		t.Fatal("deny path should reject writes")
	}
}

func TestPolicyNetworkAllowlist(t *testing.T) {
	policy, err := NewPolicy(Config{
		Mode: ModeWorkspaceWrite,
		CWD:  t.TempDir(),
		Network: NetworkPolicy{
			Enabled:        true,
			AllowedDomains: []string{"api.example.com", "*.trusted.test"},
		},
	})
	if err != nil {
		t.Fatalf("new policy: %v", err)
	}
	if err := policy.CheckNetworkURL("https://api.example.com/v1"); err != nil {
		t.Fatalf("exact domain should be allowed: %v", err)
	}
	if err := policy.CheckNetworkURL("https://assets.trusted.test/file"); err != nil {
		t.Fatalf("wildcard domain should be allowed: %v", err)
	}
	if err := policy.CheckNetworkURL("https://evil.example.test"); err == nil {
		t.Fatal("non-allowlisted domain should be rejected")
	}
}
