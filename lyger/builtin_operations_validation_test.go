package Lyger

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/l-programming-language/l/lerrors"
)

func TestBuiltinProcessLogemeCompositionOrder(t *testing.T) {
	const binding = `OB{|Key: ~<@Invocation@>~, Value: ~<@Receiver@>.wait().await?~|}`
	const guard = `BOG^GI^(E{|Statement: ~<@Invocation@>.claim_continuation()?~|}*<@Body@>)`
	for _, tc := range []struct {
		name, operation, definition, context string
	}{
		{"flat await", "await", `<DEF>` + binding + `*` + guard + `<DEF>`, "2 ordered expressions"},
		{"split await words", "await", `<DEF>[["` + binding + `"],["` + guard + `"]]<DEF>`, "one ordered word"},
		{"extra await expression", "await", `<DEF>[["` + binding + `","` + guard + `","BOD^E{|Statement: ~{}~|}"]]<DEF>`, "2 ordered expressions"},
		{"extra run expression", "run", `<DEF>[["BOD^E{|Statement: ~<@Receiver@>.run()?;~|}","BOD^E{|Statement: ~{}~|}"]]<DEF>`, "1 ordered expressions"},
		{"guard before invocation", "await", `<DEF>[["` + guard + `","` + binding + `"]]<DEF>`, "Body"},
		{"body before guard", "await", `<DEF>[["` + binding + `*<@Body@>","BOG^GI^(E{|Statement: ~<@Invocation@>.claim_continuation()?~|}*W^BOD^E{|Statement: ~{}~|})"]]<DEF>`, "Body"},
		{"duplicate continuation body", "await", `<DEF>[["` + binding + `","BOG^GI^(E{|Statement: ~<@Invocation@>.claim_continuation()?~|}*<@Body@>*<@Body@>)"]]<DEF>`, "Body"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := builtinProcessFilesystem(t)
			name := builtinProcessPath(tc.operation)
			files[name] = &fstest.MapFile{Data: []byte(tc.definition)}
			catalog, err := loadBuiltinProcesses(files)
			if catalog != nil || !errors.Is(err, lerrors.InvalidConfiguration) || !strings.Contains(err.Error(), name) || !strings.Contains(err.Error(), tc.context) {
				t.Fatalf("invalid builtin topology must reject with classified resource/role context: catalog=%v, error=%v", catalog, err)
			}
		})
	}
}
