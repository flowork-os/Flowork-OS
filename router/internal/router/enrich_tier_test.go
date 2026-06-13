package router

import "testing"

func TestIsCrewLightModel(t *testing.T) {
	if !isCrewLightModel("claude-haiku-4-5") { t.Fatal("haiku harusnya crew (skip)") }
	if isCrewLightModel("claude-sonnet-4-6") { t.Fatal("sonnet harusnya komandan (full)") }
	if isCrewLightModel("") { t.Fatal("kosong = jangan skip") }
	t.Setenv("FLOW_ROUTER_LIGHT_MODELS", "mini,nano")
	if isCrewLightModel("claude-haiku-4-5") { t.Fatal("env override: haiku ga di-list lagi") }
	if !isCrewLightModel("gpt-4o-mini") { t.Fatal("env override: mini harusnya crew") }
}
