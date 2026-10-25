package Logiceme

import (
	"reflect"
	"testing"
)

func TestParseSyllableDelimiters(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		want        []string
	}{
		{"simple", "<|S|>`A`</S|>", []string{"1", "A", "1", "false"}},
		{"expression", "<|S|>`FW{|Key: ~run~|}^W`</S|>", []string{"1", "FW{|Key: ~run~|}^W", "1", "false"}},
		{"modifiers", "<|S|>2(`E`)3</S|>", []string{"2", "E", "3", "false"}},
		{"negative modifiers", "<|S|>-12(`E`)-2</S|>", []string{"12", "E", "-2", "true"}},
		{"outer whitespace", "\n <|S|>`W`</S|>\t", []string{"1", "W", "1", "false"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseSyllable([]byte(tc.input))
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ParseSyllable = %v, %v; want %v", got, err, tc.want)
			}
		})
	}
	for _, input := range []string{
		"", "<|S|>A</S|>", "<|S|>`A</S|>", "<|S|>A`</S|>", "<|S|>``</S|>",
		"<|S|>`A`", "<|S|>2(`A`)</S|>", "<|S|>2(A)3</S|>",
		"<|S|>x(`A`)3</S|>", "<|S|>2(`A`)x</S|>",
		"<|S|>999999999999999999999999999999(`A`)1</S|>",
	} {
		t.Run(input, func(t *testing.T) {
			if _, err := ParseSyllable([]byte(input)); err == nil {
				t.Fatalf("malformed syllable %q was accepted", input)
			}
		})
	}
}

func TestParseMetaJSONStringsAndOrder(t *testing.T) {
	input := `{"AppMeta": {"Label":"a \"quote\", colon: slash\\ and\nline","Installs":"&[\"one\",\"two\"]&","AppGenDir":"/source"}}`
	got, err := ParseMetaJSON("AppMeta", input)
	if err != nil {
		t.Fatal(err)
	}
	want := []MetaObj{
		{Key: "Label", Value: "a \"quote\", colon: slash\\ and\nline", Type: "String"},
		{Key: "Installs", Value: `&["one","two"]&`, Type: "Array"},
		{Key: "AppGenDir", Value: "/source", Type: "String"},
	}
	if !reflect.DeepEqual(got.GetMetaData(), want) {
		t.Fatalf("metadata = %#v; want %#v", got.GetMetaData(), want)
	}
	if convenience := LoadMetaJSON("AppMeta", input); !reflect.DeepEqual(convenience, got) {
		t.Fatalf("LoadMetaJSON differs from checked parser")
	}
	var appended MetaDict
	appended.Add("Existing", "preserved", "String")
	appended.LoadJSON("AppMeta", input)
	if len(appended.Metas) != 4 || appended.GetVal("Existing") != "preserved" || appended.GetVal("AppGenDir") != "/source" {
		t.Fatalf("LoadJSON did not preserve and append metadata: %#v", appended.GetMetaData())
	}
	bare, err := ParseMetaJSON("", `{"Name":"module"}`)
	if err != nil || bare.GetVal("Name") != "module" {
		t.Fatalf("bare metadata = %#v, %v", bare, err)
	}
	for _, invalid := range []string{
		`{"AppMeta":`, `{"AppMeta":null}`, `{"AppMeta":[]}`, `null`,
		`{"AppMeta":{"Port":42}}`, `{"AppMeta":{"Label":null}}`, `{"AppMeta":{"Label":"ok"}} trailing`,
	} {
		if _, err := ParseMetaJSON("AppMeta", invalid); err == nil {
			t.Errorf("invalid metadata accepted: %s", invalid)
		}
		if fallback := LoadMetaJSON("AppMeta", invalid); len(fallback.Metas) != 0 {
			t.Errorf("convenience API returned partial metadata: %#v", fallback)
		}
	}
}

func TestScanJSONExactTags(t *testing.T) {
	input := "<OUTLINE>\n<JSON>\n {\"Label\":\"literal </JSON> and <JSON> text\"}\n</JSON>\n<JSON>{\"Objects\": []}</JSON>\n</OUTLINE>"
	got, err := ScanJSON(nil, []byte(input))
	if err != nil {
		t.Fatal(err)
	}
	want := [][]byte{[]byte(`{"Label":"literal </JSON> and <JSON> text"}`), []byte(`{"Objects": []}`)}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("JSON records = %q; want %q", got, want)
	}
	for _, invalid := range []string{
		`<JSON>{}</JSON`, `<JSON>{}`, `</JSON>`, `<JSON>{broken}</JSON>`,
		`<JSON></JSON>`, `<JSON>{} {}</JSON>`, `<JSON wrong>{}</JSON>`,
	} {
		if _, err := ScanJSON(nil, []byte(invalid)); err == nil {
			t.Errorf("invalid JSON markup accepted: %s", invalid)
		}
	}
}

func TestScanOutlineWhitespaceAndErrors(t *testing.T) {
	input := `<OUTLINE>
<JSON>
{"AppMeta" : {"AppGenDir" : "/source"}}
</JSON>
<JSON>{"Objects": [{"Name":"worker","Location":"/source/"}]}</JSON>
<JSON>{ "Services" : [{"Name":"root","GenericObjects":["worker"]}] }</JSON>
</OUTLINE>`
	got, err := ScanOutline(nil, []byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if got.Meta.GetVal("AppGenDir") != "/source" || len(got.Objects) != 1 || got.Objects[0].Name != "worker" || len(got.Services) != 1 {
		t.Fatalf("outline data was lost: %#v", got)
	}
	for _, invalid := range []string{
		`<JSON>{"Objects": {}}</JSON>`, `<JSON>{"AppMeta":{"Name":true}}</JSON>`, `<JSON>{"Services": [}</JSON>`,
	} {
		if _, err := ScanOutline(nil, []byte(invalid)); err == nil {
			t.Errorf("invalid outline accepted: %s", invalid)
		}
	}
	tags, err := LoadTags()
	if err != nil {
		t.Fatal(err)
	}
	program := ScannedProgram{Tags: tags}
	if _, err := program.ScanBook([]byte(`<BOOK "sample"><SETTINGS><JSON>{"key":42}</JSON></SETTINGS></BOOK>`)); err == nil {
		t.Fatal("invalid settings type was accepted")
	}
}
