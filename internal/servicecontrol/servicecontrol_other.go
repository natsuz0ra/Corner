//go:build !darwin

package servicecontrol

import kservice "github.com/kardianos/service"

type kardianosBackend struct {
	service kservice.Service
}

func newServiceBackend(prg *program, cfg *kservice.Config) (serviceBackend, error) {
	svc, err := kservice.New(prg, cfg)
	if err != nil {
		return nil, err
	}
	return &kardianosBackend{service: svc}, nil
}

func (b *kardianosBackend) Install() error {
	return b.service.Install()
}

func (b *kardianosBackend) Start() error {
	return b.service.Start()
}

func (b *kardianosBackend) Stop() error {
	return b.service.Stop()
}

func (b *kardianosBackend) Restart() error {
	return b.service.Restart()
}

func (b *kardianosBackend) Status() (string, error) {
	status, err := b.service.Status()
	if err != nil {
		return "", err
	}
	switch status {
	case kservice.StatusRunning:
		return "running", nil
	case kservice.StatusStopped:
		return "stopped", nil
	default:
		return "unknown", nil
	}
}

func (b *kardianosBackend) Uninstall() error {
	return b.service.Uninstall()
}

func (b *kardianosBackend) Run() error {
	return b.service.Run()
}
