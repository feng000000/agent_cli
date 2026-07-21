package runtime

import "myagent/pkg/llm"


type LoadSkillTool struct{}

func (t *LoadSkillTool) Name() string {
	return "load_skill"
}

func (t *LoadSkillTool) Desc() string {
	return "Load skill content"
}

func (t *LoadSkillTool) Definition() *llm.Tool {
	return &llm.Tool{
		Type: llm.ToolTypeFunction,
		Function: llm.FunctionTool{
			Name:        t.Name(),
			Description: t.Desc(),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Skill name",
					},
				},
			},
		},
	}
}

func (t *LoadSkillTool) Execute(arg string) string {
	// TODO: implement load_skill
	return "<Skill:load_skill> not implemented"
}




type ReleaseSkillTool struct{}

func (t *ReleaseSkillTool) Name() string {
	return "release_skill"
}

func (t *ReleaseSkillTool) Desc() string {
	return "Release not used skill"
}

func (t *ReleaseSkillTool) Definition() *llm.Tool {
	return &llm.Tool{
		Type: llm.ToolTypeFunction,
		Function: llm.FunctionTool{
			Name:        t.Name(),
			Description: t.Desc(),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Skill name",
					},
				},
			},
		},
	}
}

func (t *ReleaseSkillTool) Execute(arg string) string {
	// TODO: implement release_skill
	return "<Skill:release_skill> not implemented"
}
