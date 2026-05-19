package migration

import (
	"context"
	"strings"
	"testing"
)

func TestRunnerRejectsCombinedRollback(t *testing.T) {
	err := (Runner{}).Down(context.Background(), TargetAll)
	if err == nil || !strings.Contains(err.Error(), "combined rollback is not supported") {
		t.Fatalf("expected combined rollback rejection, got %v", err)
	}
}
