package contracts

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type prefixFailureWriter struct{ output bytes.Buffer }

func (writer *prefixFailureWriter) Write(data []byte) (int, error) {
	count := len(data)
	if count > 2 {
		count = 2
	}
	_, err := writer.output.Write(data[:count])
	if err != nil {
		return 0, err
	}
	return count, io.ErrShortWrite
}

func TestFailureFootprints(t *testing.T) {
	resource := ResourceRef{ID: "stdout", Kind: "external"}
	footprint := FailureFootprint{Locations: []FailureLocation{{Resource: resource}}, Observation: "operation", Outcomes: []Outcome{ReturnedFailure}}
	var writer prefixFailureWriter
	written, err := writer.Write([]byte("payload"))
	if written != 2 || !errors.Is(err, io.ErrShortWrite) || writer.output.String() != "pa" {
		t.Fatalf("fault injection did not emit prefix: %d, %v, %q", written, err, writer.output.String())
	}
	summary := FailureSummary{Outcome: ReturnedFailure, Observation: "operation", Coverage: Known, FailureChannel: "returned io error"}
	if written > 0 {
		summary.Changes = []FailureChange{{Location: FailureLocation{Resource: resource}, Partial: true, Definite: true}}
	}
	t.Run("BASE-034", func(t *testing.T) {
		if got := CheckFailurePreservation(summary, footprint, nil); got.Status != Contradicted {
			t.Fatalf("partial IO became transactional: %+v", got)
		}
		if got := CheckRetryFailureBeforeEffect(summary, footprint, nil); got.Status == Established {
			t.Fatalf("partial IO became retry-safe: %+v", got)
		}
	})
	t.Run("BASE-035", func(t *testing.T) {
		memory := FailureFootprint{Locations: []FailureLocation{{Resource: ResourceRef{ID: "memory", Kind: "region"}}}, Observation: "operation", Outcomes: []Outcome{ReturnedFailure}}
		if got := CheckFailurePreservation(summary, memory, nil); got.Status != Unresolved {
			t.Fatalf("different names became disjoint: %+v", got)
		}
		if got := CheckFailurePreservation(summary, memory, []FailureResourcePair{{Left: "memory", Right: "stdout"}}); got.Status != Established {
			t.Fatalf("proven unaffected memory lost: %+v", got)
		}
		memory.Observation = "session-end"
		if got := CheckFailurePreservation(summary, memory, []FailureResourcePair{{Left: "memory", Right: "stdout"}}); got.Status != Unresolved {
			t.Fatalf("observation boundaries mixed: %+v", got)
		}
	})
	t.Run("BASE-037", func(t *testing.T) {
		safe := summary
		safe.Changes = nil
		if got := CheckRetryFailureBeforeEffect(safe, footprint, nil); got.Status != Established {
			t.Fatal(got)
		}
		safe.Coverage = Unknown
		if got := CheckRetryFailureBeforeEffect(safe, footprint, nil); got.Status != Unresolved {
			t.Fatalf("error return alone proved retry: %+v", got)
		}
	})
	t.Run("FALS-002", func(t *testing.T) {
		safe := summary
		safe.Changes = nil
		if got := CheckFailurePreservation(safe, footprint, nil); got.Status != Established {
			t.Fatal(got)
		}
		safe.Changes = []FailureChange{{Location: footprint.Locations[0], Definite: true}}
		if got := CheckFailurePreservation(safe, footprint, nil); got.Status != Contradicted {
			t.Fatalf("injected write not refuted: %+v", got)
		}
	})
}

func TestCleanupFailurePolicy(t *testing.T) {
	t.Run("BASE-036-model", func(t *testing.T) {
		cleanup := []CleanupFailure{{ID: "second", Message: "second cleanup failure", LaunchOrder: 2}, {ID: "first", Message: "first cleanup failure", LaunchOrder: 1}, {ID: "first", Message: "duplicate observer", LaunchOrder: 1}}
		report := RuntimeCleanupReport("primary failure", 1, cleanup)
		if report.Primary != "primary failure" || report.ExitStatus != 1 || len(report.Secondary) != 2 || report.Secondary[0].ID != "first" || report.Coverage != Known {
			t.Fatalf("cleanup handling changed: %+v", report)
		}
		if got := RuntimeCleanupReport("", 0, cleanup); got.ExitStatus != 1 {
			t.Fatalf("cleanup failure preserved success: %+v", got)
		}
		if got := RuntimeCleanupReport("accepted signal", 143, cleanup); got.ExitStatus != 143 {
			t.Fatalf("cleanup replaced accepted terminal status: %+v", got)
		}
	})
}
