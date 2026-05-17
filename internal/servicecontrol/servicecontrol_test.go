package servicecontrol

import (
	"reflect"
	"testing"
)

func TestServiceConfigUsesUserService(t *testing.T) {
	cfg := serviceConfig()

	if cfg.Name != serviceName {
		t.Fatalf("service name = %q, want %q", cfg.Name, serviceName)
	}
	if !reflect.DeepEqual(cfg.Arguments, []string{"service", "run"}) {
		t.Fatalf("arguments = %#v, want service run", cfg.Arguments)
	}
	if cfg.Executable == "" {
		t.Fatal("executable should not be empty")
	}
	if got, ok := cfg.Option["UserService"].(bool); !ok || !got {
		t.Fatalf("UserService option = %#v, want true", cfg.Option["UserService"])
	}
}
