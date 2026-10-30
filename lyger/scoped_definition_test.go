package Lyger

import (
	"fmt"
	"strings"
	"testing"
)

func TestScopedDefinitionBody(t *testing.T) {
	for _, increment := range []int{1, 3} {
		t.Run(fmt.Sprint(increment), func(t *testing.T) {
			source := "Use Storage from `storage`.\n" +
				"Create \"Words\" as a Dictionary.\n" +
				"Define \"Repeat|-s|-ed|-ing|-er\" to `<DEF>BOP^X{|Iterable: ~[1, 2]~, RangeArity: ~0~|}^(<@Body@>)<DEF>`.\n" +
				fmt.Sprintf("Define \"Increment|-s|-ed|-ing|-er\" to `<DEF>DO{|Key: ~<@Value@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: +, Operands: &[<@Value@>, %d.0]&}Op]&~|}<DEF>`.\n", increment) +
				"Define \"Multiply|--ies|--ied|-ing|--ier\" to `<DEF>DO{|Key: ~<@Value@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: *, Operands: &[<@Value@>, <@Factor@>]&}Op]&~|}<DEF>`.\n\n" +
				"    Publicly create Words's \"Probe|-s|-d|--ing|--er\" as a Generic `probe` Function with No Inputs and Outputs of `f64`.\n" +
				"Create \"Count\" as a New Storage.Mutable Storage's Binding of 0.\n" +
				"Repeat a New Process for Incrementing Count and then Multiplying Count and 2.\n" +
				"Return Count.\n"
			options := canonicalTestOptions()
			observed := ""
			if increment == 3 {
				files := dictionaryIndexFilesystem(t)
				core := files["builtins/eng/core.l"]
				core.Data = []byte(strings.Replace(string(core.Data), "W^<@Definitions|Syllables@>", "W^(MOB{|Key: ~_~|}^KY{|Key: ~fetch_add~, Receiver: ~crate::OBSERVED~, Inputs: ~&[1, std::sync::atomic::Ordering::SeqCst]&~|}*<@Definitions|Syllables@>)", 1))
				options.Resources = files
				observed = "assert_eq!(OBSERVED.load(std::sync::atomic::Ordering::SeqCst),3);"
			}
			c := NewLygerService(options).NewCompilation([]byte(source), "scoped.l")
			c.Make()
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			observer := fmt.Sprintf(`#[path="src/scoped.rs"] mod generated; static OBSERVED: std::sync::atomic::AtomicUsize=std::sync::atomic::AtomicUsize::new(0); fn main(){assert_eq!(generated::probe(), %d.0);%s}`, 6*increment, observed)
			testNativeDefinitionLibrary(t, c, "Words", "scoped", observer)
		})
	}
}
