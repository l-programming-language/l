package distribution

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"
)

// AlternateRuntime supplies a test-owned L policy in a complete distribution.
// It changes App misuse presentation, retaining the original message, source
// attribution, lock lifetime and error construction. No native policy is added.
func AlternateRuntime(t testing.TB) fstest.MapFS {
	t.Helper()
	files := Files(t)
	const entry = "stdlib/eng/runtime/dictionary.l"
	const selected = "stdlib/eng/runtime/alternate.l"
	if files[entry] == nil {
		t.Fatalf("selected Runtime source is absent: %s", entry)
	}
	files[selected] = &fstest.MapFile{Data: bytes.Clone(files[entry].Data)}
	delete(files, entry)
	if bytes.Count(files["dictionaries.json"].Data, []byte(entry)) != 1 {
		t.Fatal("Runtime registry entry is not unique")
	}
	files["dictionaries.json"].Data = bytes.Replace(files["dictionaries.json"].Data, []byte(entry), []byte(selected), 1)
	const app = "stdlib/eng/runtime/app.l"
	source := string(files[app].Data)
	const start = "    Create Runtime's App's \"Describe|-s|-d|--ing|-r\""
	const end = "    Create Runtime's App's \"Explain|-s|-ed|-ing|-er\""
	first, last := strings.Index(source, start), strings.Index(source, end)
	if first < 0 || last <= first || strings.Count(source, start) != 1 || strings.Count(source, end) != 1 {
		t.Fatal("alternate policy replacement must select one complete App.Describe Function")
	}
	files[app].Data = []byte(source[:first] + alternateDescription + source[last:])
	return files
}

const alternateDescription = "    Create Runtime's App's \"Describe|-s|-d|--ing|-r\" as a Generic `app_error_description` Function with Inputs of \"value\" as `&App` and \"message\" as `&str` and Outputs of Runtime's Error.\n" + `Create "Guard" as a New Locker of a New Runtime's App's Admission of Value.
Create "Copied" as a New Storage's Copier of a New Runtime's App's Label of Guard.
Storage.Release Guard.
Create "Buffer" as a New Storage.Mutable Binding of a New Storage's Text's Owner of Message.
Create "Empty" as a New Runtime's Vacancy of Copied.
If Empty is false, Create "Target" as a New Storage.Editable Storage's Borrower of Buffer and then Create "Opening" as a New Storage's Text's Bytes of ` + "` (source: `" + ` and then Create "Addedopening" as a New Storage's Bytes's Extender of Target and Opening and then Create "Source" as a New Storage's Text's Bytes of Copied and then Create "Addedsource" as a New Storage's Bytes's Extender of Target and Source and then Create "Closing" as a New Storage's Text's Bytes of ` + "`)`" + ` and then Create "Addedclosing" as a New Storage's Bytes's Extender of Target and Closing.
Return a New Runtime's Error's Former of Buffer.

`
