package tools

import (
	"slimebot/internal/domain"
	llmsvc "slimebot/internal/services/llm"
)

// SpecialToolOptions controls which stable-name special tools are exposed.
type SpecialToolOptions struct {
	Skills             []domain.Skill
	IncludeRunSubagent bool
	IncludePlanTools   bool
}

// BuildSpecialToolDefs returns stable-name special tool definitions in a fixed order.
func BuildSpecialToolDefs(opts SpecialToolOptions) []llmsvc.ToolDef {
	var defs []llmsvc.ToolDef
	if def := BuildActivateSkillToolDef(opts.Skills); def != nil {
		defs = append(defs, *def)
	}
	if opts.IncludeRunSubagent {
		defs = append(defs, BuildRunSubagentToolDef())
	}
	if opts.IncludePlanTools {
		defs = append(defs, BuildPlanToolDefs()...)
	}
	return defs
}
