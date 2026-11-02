package Lyger

import (
	"encoding/json"
	"strings"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

// This independent native consumer composes selected production words and
// observes their concrete representation. It creates no native trust witness.
func TestContractsPointerBorrowAndMetadataNative(t *testing.T) {
	words := []string{
		`MOB{|Key: ~bytes~|}^KY{|Key: ~to_vec~, Receiver: ~b"original"~|}`,
		`MOB{|Key: ~signature~|}^KY{|Key: ~Vec::from~, Inputs: ~&[[2]]&~|}`,
		`MOB{|Key: ~value~|}^KY{|Key: ~l_value_wrap~, Inputs: ~&[bytes, signature, l_show_text, l_reach_scalar]&~|}`,
		`MOB{|Key: ~element~|}^KY{|Key: ~Vec::from~, Inputs: ~&[[2]]&~|}`,
		`MOB{|Key: ~pointer~|}^KY{|Key: ~l_pointer_new~, Inputs: ~&[element, value]&~|}`,
		`OB{|Key: ~pointer~, Value: ~pointer?~|}`,
		`MOB{|Key: ~copy~|}^KY{|Key: ~<@PointerCopy@>~, Inputs: ~&[&pointer]&~|}`,
		`MOB{|Key: ~_~|}^KY{|Key: ~assert_eq!~, Inputs: ~&[&pointer.element, &copy.element]&~|}`,
		`MOB{|Key: ~original_metadata~|}^KY{|Key: ~as_ptr~, Receiver: ~pointer.element~|}`,
		`MOB{|Key: ~copied_metadata~|}^KY{|Key: ~as_ptr~, Receiver: ~copy.element~|}`,
		`MOB{|Key: ~_~|}^KY{|Key: ~assert_ne!~, Inputs: ~&[original_metadata, copied_metadata]&~|}`,
		`MOB{|Key: ~shared~|}^KY{|Key: ~std::rc::Rc::ptr_eq~, Inputs: ~&[&pointer.target, &copy.target]&~|}`,
		`MOB{|Key: ~_~|}^KY{|Key: ~assert!~, Inputs: ~&[shared]&~|}`,
		`MOB{|Key: ~guard~|}^KY{|Key: ~borrow_mut~, Receiver: ~pointer.target~|}`,
		`MOB{|Key: ~read~|}^KY{|Key: ~l_pointer_get~, Inputs: ~&[&copy]&~|}`,
		`MOB{|Key: ~read_failure~|}^KY{|Key: ~err~, Receiver: ~read~|}`,
		`MOB{|Key: ~read_failure~|}^KY{|Key: ~expect~, Receiver: ~read_failure~, Inputs: ~&["held borrow must reject read"]&~|}`,
		`MOB{|Key: ~read_message~|}^KY{|Key: ~to_string~, Receiver: ~read_failure~|}`,
		`MOB{|Key: ~_~|}^KY{|Key: ~assert_eq!~, Inputs: ~&[read_message, "Pointer is already borrowed"]&~|}`,
		`MOB{|Key: ~bytes~|}^KY{|Key: ~to_vec~, Receiver: ~b"rejected"~|}`,
		`MOB{|Key: ~signature~|}^KY{|Key: ~Vec::from~, Inputs: ~&[[2]]&~|}`,
		`MOB{|Key: ~value~|}^KY{|Key: ~l_value_wrap~, Inputs: ~&[bytes, signature, l_show_text, l_reach_scalar]&~|}`,
		`MOB{|Key: ~write~|}^KY{|Key: ~l_pointer_set~, Inputs: ~&[&copy, value]&~|}`,
		`MOB{|Key: ~write_failure~|}^KY{|Key: ~err~, Receiver: ~write~|}`,
		`MOB{|Key: ~write_failure~|}^KY{|Key: ~expect~, Receiver: ~write_failure~, Inputs: ~&["held borrow must reject write"]&~|}`,
		`MOB{|Key: ~write_message~|}^KY{|Key: ~to_string~, Receiver: ~write_failure~|}`,
		`MOB{|Key: ~_~|}^KY{|Key: ~assert_eq!~, Inputs: ~&[write_message, "Pointer is already borrowed"]&~|}`,
		`MOB{|Key: ~_~|}^KY{|Key: ~std::mem::drop~, Inputs: ~&[guard]&~|}`,
		`MOB{|Key: ~unchanged~|}^KY{|Key: ~l_pointer_get~, Inputs: ~&[&pointer]&~|}`,
		`OB{|Key: ~unchanged~, Value: ~unchanged?~|}`,
		`MOB{|Key: ~bytes~|}^KY{|Key: ~l_display~, Inputs: ~&[&unchanged]&~|}`,
		`OB{|Key: ~bytes~, Value: ~bytes?~|}`,
		`MOB{|Key: ~_~|}^KY{|Key: ~assert_eq!~, Inputs: ~&[bytes, b"original"]&~|}`,
		`MOB{|Key: ~success~|}^KY{|Key: ~Ok~, Inputs: ~&[()]&~|}`,
		`BOD^E{|Statement: ~success~|}`,
	}
	encoded, err := json.Marshal([][]string{words})
	if err != nil {
		t.Fatal(err)
	}
	// Source framing decodes its documented backslash escapes before JSON.
	grouped := strings.ReplaceAll(string(encoded), `\`, `\\`)
	source := "Define \"Contractpointerobservation\" as `<DEF>" + grouped + "<DEF>`.\nCreate \"Witness\" as a New Basic App.\n"
	compilation := requireDistributionFrontend(t, source, nil).Compilation
	var declarations []DefinitionInstance
	for _, name := range []string{"Ignore", "Value.Wrap", "Value.Format", "Pointer.Create", "Pointer.Read", "Pointer.Update", "Pointer.Copy", "Text.Render"} {
		record, err := compilation.builtinRecord("Storage." + name)
		if err != nil {
			t.Fatal(err)
		}
		declaration, err := compilation.composeExport(record)
		if err != nil {
			t.Fatal(err)
		}
		declarations = append(declarations, declaration)
	}
	emitter := orchestrationEmitter{compilation: compilation}
	pointerCopy, err := emitter.functionReference("Storage.Pointer.Copy")
	if err != nil {
		t.Fatal(err)
	}
	declarations, err = definitionClosure(append(declarations, emitter.dependencies...)...)
	if err != nil {
		t.Fatal(err)
	}
	binding, state, err := compilation.Scope.Root.lookupLocal([]byte("Contractpointerobservation"))
	if err != nil || state != lookupFound || binding == nil {
		t.Fatalf("native observer definition: %v", err)
	}
	template, err := compilation.recordTemplate(binding.Record)
	if err != nil || template == nil {
		t.Fatalf("native observer template: %v", err)
	}
	observation, err := template.Instantiate(map[string]DefinitionArgument{"PointerCopy": pointerCopy})
	if err != nil {
		t.Fatal(err)
	}
	var body []string
	for _, word := range observation.Words {
		body = append(body, word...)
	}
	entry, err := compilation.futureEntry("l_main", "runtime::Result<()>", body)
	if err != nil {
		t.Fatal(err)
	}
	declarations = append(declarations, entry)
	main := newRootObject("main")
	primary := newRootService("main")
	primary.Objects = []*object{main}
	app := &App{ID: "entry", Meta: appMetadata("entry"), Services: []*service{primary}}
	word := &w{Syllables: []*s{{CompiledBytes: []byte("A")}}}
	for _, declaration := range declarations {
		word.Syllables = append(word.Syllables, &s{CompiledBytes: []byte(declaration.Expression)})
	}
	program := &Program{Books: []*book{{Name: "entry", Chapters: []*chapter{{Name: "main", sentences: []*sentence{{Name: "main", Words: []*w{word}}}}}}}}
	_, aml, err := renderProgram(program, []*App{app})
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate(aml, Logiceme.GenerationOptions{IncludeRuntime: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("generation: actual selected storage definitions and typed Pointer.Copy call composed through AML")
	recordBundle(t, bundle)
	binary := buildActualBundle(t, bundle, "entry", true, "")
	t.Log("native compilation: actual returned bundle and retained Cargo manifest/lock compiled")
	stdout, stderr, err := nativetest.RuntimeExecution.Run(t, binary, "", nil)
	if err != nil || len(stdout) != 0 || len(stderr) != 0 {
		t.Fatalf("native borrow/metadata observer: error=%v stdout=%q stderr=%q", err, stdout, stderr)
	}
	t.Log("runtime: held borrow rejects read/write and preserves value; Pointer.Copy duplicates nonempty signature allocation and retains shared target")
}
