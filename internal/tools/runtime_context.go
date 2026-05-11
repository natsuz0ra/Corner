package tools

import (
	"context"

	skillsvc "slimebot/internal/services/skill"
)

type skillRuntimeContextKey struct{}
type todoStateContextKey struct{}
type processManagerContextKey struct{}

func WithSkillRuntime(ctx context.Context, runtime *skillsvc.SkillRuntimeService) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if runtime == nil {
		return ctx
	}
	return context.WithValue(ctx, skillRuntimeContextKey{}, runtime)
}

func skillRuntimeFromContext(ctx context.Context) (*skillsvc.SkillRuntimeService, bool) {
	if ctx == nil {
		return nil, false
	}
	runtime, ok := ctx.Value(skillRuntimeContextKey{}).(*skillsvc.SkillRuntimeService)
	return runtime, ok && runtime != nil
}

func WithTodoState(ctx context.Context, state *TodoState) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if state == nil {
		return ctx
	}
	return context.WithValue(ctx, todoStateContextKey{}, state)
}

func todoStateFromContext(ctx context.Context) *TodoState {
	if ctx == nil {
		return defaultTodoState
	}
	if state, ok := ctx.Value(todoStateContextKey{}).(*TodoState); ok && state != nil {
		return state
	}
	return defaultTodoState
}

func WithProcessManager(ctx context.Context, manager *ProcessManager) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if manager == nil {
		return ctx
	}
	return context.WithValue(ctx, processManagerContextKey{}, manager)
}

func processManagerFromContext(ctx context.Context) *ProcessManager {
	if ctx == nil {
		return defaultProcessManager
	}
	if manager, ok := ctx.Value(processManagerContextKey{}).(*ProcessManager); ok && manager != nil {
		return manager
	}
	return defaultProcessManager
}
