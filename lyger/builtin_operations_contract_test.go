package Lyger

import (
	"strings"
	"testing"
	"testing/fstest"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

const builtinProcessDirectory = "builtins/operations/rust/"

// These definitions are independently authored test INPUTS, not generated-output
// replacements. Default-resource tests never use this overlay. The unannotated
// slots belong to the private compiler binder, not the public argument API.
func builtinProcessDefinitions() map[string]string {
	return map[string]string{
		"build":               `<DEF>MOB{|Key: ~lProcessResult~|}^KY{|Key: ~runtime::app_build~, Imports: ~&[runtime::app_build as _]&~, Inputs: ~&[&<@Receiver@>]&~|}*OB{|Key: ~_~, Value: ~lProcessResult?~|}<DEF>`,
		"run":                 `<DEF>MOB{|Key: ~lProcessResult~|}^KY{|Key: ~runtime::app_run~, Imports: ~&[runtime::app_run as _]&~, Inputs: ~&[&<@Receiver@>]&~|}*OB{|Key: ~_~, Value: ~lProcessResult?~|}<DEF>`,
		"stop":                `<DEF>MOB{|Key: ~lProcessFuture~|}^KY{|Key: ~runtime::app_stop~, Imports: ~&[runtime::app_stop as _]&~, Inputs: ~&[&<@Receiver@>]&~|}*COB{|Key: ~lProcessResult~|}^DV{|Key: ~lProcessFuture~|}*OB{|Key: ~_~, Value: ~lProcessResult?~|}<DEF>`,
		"await":               `<DEF>[["MOB{|Key: ~lProcessFuture~|}^KY{|Key: ~runtime::app_wait~, Imports: ~&[runtime::app_wait as _]&~, Inputs: ~&[&<@Receiver@>]&~|}*COB{|Key: ~lProcessResult~|}^DV{|Key: ~lProcessFuture~|}*OB{|Key: ~<@Invocation@>~, Value: ~lProcessResult?~|}", "MOB{|Key: ~lProcessClaim~|}^KY{|Key: ~runtime::invocation_claim~, Imports: ~&[runtime::invocation_claim as _]&~, Inputs: ~&[&<@Invocation@>]&~|}*OB{|Key: ~lProcessClaimed~, Value: ~lProcessClaim?~|}*BOG^GI^(E{|Statement: ~lProcessClaimed~|}*<@Body@>)"]]<DEF>`,
		"finish-continuation": `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~runtime::invocation_finish~, Imports: ~&[runtime::invocation_finish as _]&~, Inputs: ~&[&<@Invocation@>]&~|}<DEF>`,
	}
}

func builtinProcessPath(operation string) string {
	return builtinProcessDirectory + "process-" + operation + ".def"
}

func builtinProcessFilesystem(t *testing.T) fstest.MapFS {
	t.Helper()
	files := dictionaryIndexFilesystem(t)
	for operation, definition := range builtinProcessDefinitions() {
		files[builtinProcessPath(operation)] = &fstest.MapFile{Data: []byte(definition)}
	}
	return files
}

func builtinProcessRust(bundle Logiceme.Executable) string {
	var result strings.Builder
	for i, paths := range bundle.MapArrays {
		for j, path := range paths {
			if strings.HasSuffix(path, ".rs") {
				result.Write(bundle.ByteArrays[i][j])
			}
		}
	}
	return result.String()
}
