package Lyger

import (
	"path/filepath"
	"strings"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

func TestStructuredProcessInputTypesRejectMismatch(t *testing.T) {
	for _, tc := range []struct{ name, declarations, input, declared, body string }{
		{"option-owned", "", "Option<f64>", "String", "Choose Value, a New Present using a New Process with Inputs of \"payload\" as `String` for Returning Payload, and a New Absent using a New Process for Returning 7."},
		{"option-exclusive", "", "Option<f64>", "&mut String", "Choose Value, a New Present using a New Process with Inputs of \"payload\" as `&mut String` for Returning Payload, and a New Absent using a New Process for Returning 7."},
		{"result-success", "", "Result<f64,u8>", "String", "Storage.Result.Select Value, a New Process with Inputs of \"payload\" as `String` for Returning 1, and a New Process with Inputs of \"failure\" as `u8` for Returning 0."},
		{"result-error", "", "Result<u8,f64>", "&mut String", "Storage.Result.Select Value, a New Process with Inputs of \"payload\" as `u8` for Returning 1, and a New Process with Inputs of \"failure\" as `&mut String` for Returning 0."},
		{"shared-to-exclusive", "", "Option<&String>", "&mut String", "Choose Value, a New Present using a New Process with Inputs of \"payload\" as `&mut String` for Returning 1, and a New Absent using a New Process for Returning 0."},
		{"iteration", "", "&[f64]", "&String", "Perform a New Storage's Iteration of Value and a New Process with Inputs of \"item\" as `&String` for Returning 1.\nReturn 0."},
		{"pair", "", "&[(u8,f64)]", "&String", "Walk Value and a New Process with Inputs of \"first\" as `&u8` and \"second\" as `&String` for Returning 1.\nReturn 0."},
		{"nominal-shadow", "Define \"Carrier\" as `<DEF>HY{|Key: ~NativeOriginal~, Props: ~&[]&~|}<DEF>`.\n", "Option<crate::NativeOriginal>", "NativeOther",
			"Define \"Carrier\" as `<DEF>HY{|Key: ~NativeOther~, Props: ~&[]&~|}<DEF>`.\nChoose Value, a New Present using a New Process with Inputs of \"payload\" as Carrier for Returning 1, and a New Absent using a New Process for Returning 0."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := processInputWords + tc.declarations + "Define \"Perform|-s|-ed|-ing|-er\" to `<DEF><@Body@><DEF>`.\n\n" +
				"    Publicly create \"Exercise\" as a Generic Function with Inputs of \"value\" as `" + tc.input + "` and Outputs of `f64`.\n" + tc.body + "\n"
			c := compileDefinitionSource(t, source, "process-input-type.l")
			aml, err := c.ComposeLibrary("Exercise", "process_input")
			if err != nil {
				t.Fatal(err)
			}
			bundle, err := Logiceme.NewLogicemeService(c.Gram.resources).Evaluate(aml)
			if err != nil {
				t.Fatal(err)
			}
			files := make(map[string][]byte)
			for index, name := range bundle.MapArrays[0] {
				files[strings.TrimPrefix(name, "/")] = bundle.ByteArrays[0][index]
			}
			files["observe.rs"] = []byte("pub struct NativeOriginal; #[path=\"src/process_input.rs\"] mod generated; fn main() {}")
			directory := nativetest.Materialize(t, files)
			tools := nativetest.Toolchain(t)
			_, diagnostic, err := nativetest.BuildPrerequisite.Run(t, tools.Rustc, directory, nativetest.CleanEnvironment(tools.Rustc), "--edition=2024", "observe.rs", "-o", filepath.Join(directory, "observe"))
			if err == nil {
				t.Fatalf("Process input %s silently accepted incompatible payload from %s", tc.declared, tc.input)
			}
			if !strings.Contains(string(diagnostic), "mismatched types") || !strings.Contains(string(diagnostic), tc.declared) {
				t.Fatalf("Process type rejection lost the declared type: %v: %s", err, diagnostic)
			}
		})
	}
}

func TestStructuredProcessInputOwnership(t *testing.T) {
	for _, tc := range []struct{ name, declarations, constructor, carrier, payload, observer string }{
		{"owned", "", "Function", "Option<crate::Owned>", "`crate::Owned`",
			"let value = Owned(Box::new(7)); let address = &*value.0 as *const u8; let kept = generated::exercise(Some(value)).unwrap(); assert_eq!(&*kept.0 as *const u8, address); assert_eq!(DROPS.load(Ordering::SeqCst),0); drop(kept); assert_eq!(DROPS.load(Ordering::SeqCst),1); assert!(generated::exercise(None).is_none());"},
		{"shared", "", "Function", "Option<&String>", "`&String`",
			"let value = String::from(\"shared\"); let kept = generated::exercise(Some(&value)).unwrap(); assert!(std::ptr::eq(kept,&value)); assert_eq!(value,\"shared\");"},
		{"exclusive", "", "Function", "Option<&mut String>", "`&mut String`",
			"let mut value = String::from(\"exclusive\"); let address = &value as *const String; { let kept = generated::exercise(Some(&mut value)).unwrap(); assert_eq!(kept as *const String,address); kept.push('!'); } value.push('?'); assert_eq!(value,\"exclusive!?\");"},
		{"generic", "", "Storage.Routine of `T`", "Option<T>", "`T`",
			"let value = Owned(Box::new(8)); let address = &*value.0 as *const u8; let kept = generated::exercise(Some(value)).unwrap(); assert_eq!(&*kept.0 as *const u8,address); drop(kept); assert_eq!(DROPS.load(Ordering::SeqCst),1);"},
		{"nominal-alias", "Define \"Carrier\" as `<DEF>HY{|Key: ~NativeCarrier~, Props: ~&[]&~|}<DEF>`.\nUse Carrier as \"Carriertype\".\n", "Function", "Option<NativeCarrier>", "Carriertype",
			"assert!(generated::exercise(Some(generated::NativeCarrier {})).is_some()); assert!(generated::exercise(None).is_none());"},
		{"inferred", "", "Function", "Option<crate::Owned>", "`_`",
			"let kept=generated::exercise(Some(Owned(Box::new(9)))).unwrap(); assert_eq!(*kept.0,9); drop(kept); assert_eq!(DROPS.load(Ordering::SeqCst),1);"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := "Choose Value, a New Present using a New Process with Inputs of \"payload\" as " + tc.payload + " for Returning a New Storage's Present of Payload, and a New Absent using a New Process for Returning a New Storage's Absent."
			source := processInputWords + tc.declarations + "\n    Publicly create \"Exercise\" as a Generic `exercise` " + tc.constructor + " with Inputs of \"value\" as `" + tc.carrier + "` and Outputs of `" + tc.carrier + "`.\n" + body + "\n"
			c := compileDefinitionSource(t, source, "process-input-ownership.l")
			if tc.name == "nominal-alias" {
				definition, err := c.ComposeDefinition("Exercise")
				if err != nil {
					t.Fatal(err)
				}
				labelProvenance, typeProvenance := false, false
				for _, binding := range definition.Bindings {
					for _, span := range binding.Sources {
						if span.Path == "process-input-ownership.l" && span.Start.Line == strings.Count(source[:strings.Index(source, "Choose Value")], "\n")+1 {
							labelProvenance = labelProvenance || binding.Slot.Name == "Payload" && binding.Slot.Type == "Label"
							typeProvenance = typeProvenance || binding.Slot.Type == "Type" && binding.Emitted == "NativeCarrier"
						}
					}
				}
				if !labelProvenance || !typeProvenance {
					t.Fatal("Process input label or type annotation lost the original input source span")
				}
				canonical, err := canonicalRecord(c.Memory["Carriertype"], c.Semantics.Aliases)
				if err != nil || canonical != c.Memory["Carrier"] {
					t.Fatalf("type alias lost its selected declaration: %v", err)
				}
				retained := false
				for _, rec := range c.Semantics.Subjects {
					if argument := rec.definitionArgument; argument != nil && argument.processType != nil {
						for _, dependency := range argument.processType.dependencies {
							retained = retained || dependency.definition == canonical && argument.processType.source == argument.source
						}
					}
				}
				if !retained {
					t.Fatal("Process input dropped the canonical native declaration dependency")
				}
			}
			testNativeDefinitionLibrary(t, c, "Exercise", "ownership", `use std::sync::atomic::{AtomicUsize,Ordering};
static DROPS: AtomicUsize=AtomicUsize::new(0);
pub struct Owned(Box<u8>);
impl Drop for Owned { fn drop(&mut self) { DROPS.fetch_add(1,Ordering::SeqCst); } }
#[path="src/ownership.rs"] mod generated;
fn main() { `+tc.observer+` }`)
		})
	}
}

func TestStructuredProcessIterationOwnership(t *testing.T) {
	for _, tc := range []struct{ name, values, payload, observer string }{
		{"owned", "Vec<String>", "String", "let input=String::from(\"first\"); let address=input.as_ptr(); let kept=generated::exercise(vec![input,String::from(\"second\")]).unwrap(); assert_eq!(kept.as_ptr(),address); assert_eq!(kept,\"first\");"},
		{"borrowed", "&[String]", "&String", "let values=[String::from(\"first\")]; let kept=generated::exercise(&values).unwrap(); assert!(std::ptr::eq(kept,&values[0]));"},
		{"exclusive", "&mut [String]", "&mut String", "let mut values=[String::from(\"first\")]; generated::exercise(&mut values).unwrap().push('!'); assert_eq!(values[0],\"first!\");"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := processInputWords + "Define \"Perform|-s|-ed|-ing|-er\" to `<DEF><@Body@><DEF>`.\n\n" +
				"    Publicly create \"Exercise\" as a Generic `exercise` Function with Inputs of \"values\" as `" + tc.values + "` and Outputs of `Option<" + tc.payload + ">`.\n" +
				"Perform a New Storage's Iteration of Values and a New Process with Inputs of \"item\" as `" + tc.payload + "` for Returning a New Storage's Present of Item.\nReturn a New Storage's Absent.\n"
			c := compileDefinitionSource(t, source, "process-iteration-ownership.l")
			testNativeDefinitionLibrary(t, c, "Exercise", "ownership", `#[path="src/ownership.rs"] mod generated; fn main() { `+tc.observer+` }`)
		})
	}
}

const processInputWords = "Use Storage from `storage`.\n" +
	"Define \"Choose|-s|-d|--ing|-r\" to `<DEF>BOG^GI{|Key: ~<@Value@>~|}^(<@First@>*<@Second@>)<DEF>`.\n" +
	"Define \"Present\" as `<DEF>E{|Statement: ~Some(<@Payload|Label@>)~|}*(<@Body@>)<DEF>`.\n" +
	"Define \"Absent\" as `<DEF>E{|Statement: ~None~|}*(<@Body@>)<DEF>`.\n" +
	"Define \"Wrap|-s|-ped|-ping|-per\" to `<DEF>W^<@Body@><DEF>`.\n" +
	"Define \"Walk|-s|-ed|-ing|-er\" to `<DEF>BOP^X{|Iterable: ~<@Values@>~, Key: ~<@Second|Label@>~, Index: ~<@First|Label@>~, RangeArity: ~2~|}^(<@Body@>)<DEF>`.\n"

func TestStructuredProcessInputs(t *testing.T) {
	for _, tc := range []struct{ name, inputs, output, body, observer string }{
		{"selection", "\"value\" as `Option<f64>`", "f64",
			"Choose Value, a New Present using a New Process with Inputs of \"payload\" as `f64` for Returning Payload, and a New Absent using a New Process for Returning 7.\n",
			"assert_eq!(generated::exercise(Some(3.0)), 3.0); assert_eq!(generated::exercise(None), 7.0);"},
		{"pairs", "\"values\" as `&[(u64, u64)]`", "u64",
			"Walk Values and a New Process with Inputs of \"first\" as `&u64` and \"second\" as `&u64` for Returning a New Storage's Dereference of Second.\nReturn a New Storage's Integral of 0 and `u64`.\n",
			"assert_eq!(generated::exercise(&[(2, 9), (3, 4)]), 9); assert_eq!(generated::exercise(&[]), 0);"},
		{"result", "\"value\" as `Option<f64>`", "Result<f64, f64>",
			"Choose Value, a New Present using a New Process with Inputs of \"payload\" as `f64` for Returning a New Storage's Success of Payload, and a New Absent using a New Process for Creating \"problem\" as a New Storage's Success of 7 and then Returning Problem.\n",
			"assert_eq!(generated::exercise(Some(3.0)), Ok(3.0)); assert_eq!(generated::exercise(None), Ok(7.0));"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := processInputWords + "\n    Publicly create \"Exercise\" as a Generic Function with Inputs of " + tc.inputs + " and Outputs of `" + tc.output + "`.\n" + tc.body
			c := compileDefinitionSource(t, source, "process-inputs.l")
			testNativeDefinitionLibrary(t, c, "Exercise", "scoped", `#[path="src/scoped.rs"] mod generated; fn main() { `+tc.observer+` }`)
		})
	}
}

func TestStructuredProcessInputsCannotEscape(t *testing.T) {
	for _, body := range []string{
		"Choose Value, a New Present using a New Process with Inputs of \"payload\" as `f64` for Returning Payload, and a New Absent using a New Process for Returning Payload.\n",
		"Choose Value, a New Present using a New Process with Inputs of \"payload\" as `f64` for Processing a New Storage's Discarder of Payload, and a New Absent using a New Storage.Empty Process.\nReturn Payload.\n",
	} {
		c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(processInputWords+"\n    Publicly create \"Exercise\" as a Generic Function with Inputs of \"value\" as `Option<f64>` and Outputs of `f64`.\n"+body), "escaping-input.l")
		c.Make()
		if c.Error == nil || !strings.Contains(strings.ToLower(c.Error.Error()), "payload") {
			t.Fatalf("escaped Process input: %v", c.Error)
		}
	}
}

func TestStructuredProcessBindingArity(t *testing.T) {
	for _, tc := range []struct{ name, expression, diagnostic string }{
		{"missing", "Choose Value, a New Present using a New Process for Returning 1, and a New Absent using a New Process for Returning 0.", "missing argument for Body"},
		{"extra", "Choose Value, a New Present using a New Process with Inputs of \"first\" as `f64` and \"second\" as `f64` for Returning First, and a New Absent using a New Process for Returning 0.", "inputs"},
		{"unbound", "Wrap a New Process with Inputs of \"payload\" as `f64` for Returning Payload.", "inputs"},
		{"text-label", "Choose Value, a New Present of `payload` using a New Process for Returning 1, and a New Absent using a New Process for Returning 0.", "missing argument for Body"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(processInputWords+"\n    Publicly create \"Exercise\" as a Generic Function with Inputs of \"value\" as `Option<f64>` and Outputs of `f64`.\n"+tc.expression+"\n"), "process-arity.l")
			c.Make()
			err := c.Error
			if err == nil {
				_, err = c.ComposeLibrary("Exercise", "arity")
			}
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.diagnostic)) {
				t.Fatalf("invalid Process binding accepted or wrong diagnostic: %v", err)
			}
		})
	}
}
