package contracts

import "testing"

func TestCleanupCauseIdentity(t *testing.T) {
	primary := CleanupFailure{ID: "primary", Message: "same text", LaunchOrder: 1}
	distinct := CleanupFailure{ID: "other", Message: "same text", LaunchOrder: 2}
	report := RuntimeCleanupReportWithCauses(primary.Message, []string{primary.ID}, 1, []CleanupFailure{primary, distinct, distinct})
	if report.Primary != "same text" || len(report.Secondary) != 1 || report.Secondary[0].ID != distinct.ID || report.ExitStatus != 1 {
		t.Fatalf("cleanup deduplication must follow cause identity, not message equality: %+v", report)
	}
}
