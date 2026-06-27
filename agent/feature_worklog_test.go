package main

import (
	"os"
	"testing"
)

func TestWorklogEnvHelpers(t *testing.T) {
	// worklogEnabled: default ON; "0"/"false" = OFF.
	os.Unsetenv("FLOWORK_WORKLOG")
	if !worklogEnabled() {
		t.Error("default mau ON")
	}
	os.Setenv("FLOWORK_WORKLOG", "0")
	if worklogEnabled() {
		t.Error("'0' mau OFF")
	}
	os.Unsetenv("FLOWORK_WORKLOG")

	// worklogStaleMin: default 60; valid override.
	os.Unsetenv("FLOWORK_WORKLOG_STALE_MIN")
	if worklogStaleMin() != 60 {
		t.Errorf("default mau 60, dapet %d", worklogStaleMin())
	}
	os.Setenv("FLOWORK_WORKLOG_STALE_MIN", "15")
	if worklogStaleMin() != 15 {
		t.Errorf("override mau 15, dapet %d", worklogStaleMin())
	}
	os.Unsetenv("FLOWORK_WORKLOG_STALE_MIN")

	// worklogOrchestrator: default mr-flow.
	os.Unsetenv("FLOWORK_ORCHESTRATOR")
	if worklogOrchestrator() != "mr-flow" {
		t.Errorf("default mau mr-flow, dapet %s", worklogOrchestrator())
	}
}
