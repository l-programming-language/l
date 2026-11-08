package Lyger

import (
	"os"
	"path/filepath"
	"testing"
)

// Read the published caller bytes, rather than a second copy of the examples.
// These native-signature modules require the host supplied by the observers;
// the documentation therefore classifies them as fragments, not App programs.
func TestWordDocumentationNativeCallers(t *testing.T) {
	read := func(t *testing.T, name string) *Compilation {
		t.Helper()
		file := filepath.Join("..", "docs", "examples", name)
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		return compileDefinitionSource(t, string(source), name)
	}
	t.Run("core", func(t *testing.T) {
		c := read(t, "word-core.l")
		testNativeDefinitionLibrary(t, c, "Corewords", "documented_core", `#[path="src/documented_core.rs"] mod generated;
fn main() {
 let value = std::sync::Arc::new(7_u8);
 let retained = generated::retained(&value);
 assert!(std::sync::Arc::ptr_eq(&value, &retained));
 assert_eq!(std::sync::Arc::strong_count(&value), 2);
}`)
	})
	t.Run("io", func(t *testing.T) {
		c := read(t, "word-io.l")
		testNativeDefinitionLibrary(t, c, "Writing", "documented_io", `#[path="src/documented_io.rs"] mod generated;
fn main() {
 let _: std::io::Stdout = generated::output();
 let _: std::io::Stderr = generated::diagnostic();
 let mut writer = vec![42];
 generated::send(&mut writer, &[0,255,b'\n']).unwrap();
 assert_eq!(writer, [42,0,255,b'\n']);
 generated::line(&mut writer, b"done").unwrap();
 assert_eq!(writer, [42,0,255,b'\n',b'd',b'o',b'n',b'e',b'\n']);
 generated::flush(&mut writer).unwrap();
 assert_eq!(writer, [42,0,255,b'\n',b'd',b'o',b'n',b'e',b'\n']);
}`)
	})
	t.Run("selection", func(t *testing.T) {
		c := read(t, "word-selection.l")
		wordLibraryRuntimeObserver(t, c, "Selection", `pub(crate) fn exercise() {
 assert_eq!(super::choose(Some(7)), 7);
 assert_eq!(super::choose(None), 0);
 assert_eq!(super::preserve(Ok(8)), Ok(8));
 assert_eq!(super::preserve(Err(9)), Err(9));
 let payload = Box::new(10); let address = &*payload as *const u8;
 assert_eq!(&*super::require(Some(payload), "missing").unwrap() as *const u8, address);
 assert_eq!(super::require(None, "missing").unwrap_err().to_string(), "missing");
 let payload = Box::new(11); let address = &*payload as *const u8;
 assert_eq!(&*super::expect(Ok(payload), "required") as *const u8, address);
 std::panic::set_hook(Box::new(|_| {}));
 let panic = std::panic::catch_unwind(|| super::expect(Err(12), "required")).unwrap_err();
 assert_eq!(panic.downcast_ref::<String>().unwrap(), "required: 12");
 println!("word-library error assertions reached");
}`)
	})
}
