//go:build windows

package updater

import "os/exec"

func detachProcess(_ *exec.Cmd) {}
