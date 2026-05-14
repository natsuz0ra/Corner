package tools

import (
	"context"

	skillsvc "slimebot/internal/services/skill"
)

type skillRuntimeContextKey struct{}
type activatedSkillsContextKey struct{}
type subagentRunnerContextKey struct{}
type todoStateContextKey struct{}
type processManagerContextKey struct{}

// SubagentRunRequest is the tool-layer input passed to the chat agent runner.
type SubagentRunRequest struct {
	Title   string
	Task    string
	Context string
	Params  map[string]any
}

// SubagentRunner executes a nested agent for run_subagent.
type SubagentRunner interface {
	RunSubagent(ctx context.Context, request SubagentRunRequest) (*ExecuteResult, error)
}

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

func WithActivatedSkills(ctx context.Context, activated map[string]struct{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if activated == nil {
		return ctx
	}
	return context.WithValue(ctx, activatedSkillsContextKey{}, activated)
}

func activatedSkillsFromContext(ctx context.Context) map[string]struct{} {
	if ctx == nil {
		return map[string]struct{}{}
	}
	if activated, ok := ctx.Value(activatedSkillsContextKey{}).(map[string]struct{}); ok && activated != nil {
		return activated
	}
	return map[string]struct{}{}
}

func WithSubagentRunner(ctx context.Context, runner SubagentRunner) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if runner == nil {
		return ctx
	}
	return context.WithValue(ctx, subagentRunnerContextKey{}, runner)
}

func subagentRunnerFromContext(ctx context.Context) (SubagentRunner, bool) {
	if ctx == nil {
		return nil, false
	}
	runner, ok := ctx.Value(subagentRunnerContextKey{}).(SubagentRunner)
	return runner, ok && runner != nil
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
