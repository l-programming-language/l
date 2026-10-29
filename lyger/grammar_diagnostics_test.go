package Lyger

import "testing"

type rejectedGrammarDiagnostic struct{ rendered int }

func (d *rejectedGrammarDiagnostic) Error() string {
	d.rendered++
	return "rejected alternative"
}

func TestSuccessfulGrammarDoesNotRenderRejectedDiagnostics(t *testing.T) {
	diagnostic := &rejectedGrammarDiagnostic{}
	tokens := []*token{{rawBytes: []byte("value")}}
	state := 0
	err := selectGrammarCase(&state, map[string]string{"a": "reject", "b": "accept"},
		func(v int) int { return v },
		func(v *int, rule string) error {
			if rule == "reject" {
				return applyGrammarRule(v, "ID", func(v int) int { return v }, nil,
					func(*int, grammarTerm) error { return diagnostic }, nil)
			}
			*v = 1
			return nil
		}, func(v *int) []*token { return tokens[*v:] }, func(*int, string) {}, tokens)
	if err != nil || state != 1 || diagnostic.rendered != 0 {
		t.Fatalf("selected state=%d, error=%v; rendered %d discarded diagnostics", state, err, diagnostic.rendered)
	}
}
