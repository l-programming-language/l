package Logiceme

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/l-programming-language/l/lerrors"
)

const runtimeCompositionDirectory = "builtins/runtime"

// runtimeComposition describes derived AML, never native source. Its source
// fingerprints make a partial or stale resource distribution fail closed. The
// build tool regenerates the AML from those source words outside Logiceme.
type runtimeComposition struct {
	Version int               `json:"version"`
	SHA256  string            `json:"sha256"`
	Sources map[string]string `json:"sources"`
}

// Keep one successful rendering, up to 1 MiB. Providers are reread and validated
// before lookup; no filesystem identity, mutable parsed graph, or error is cached.
var runtimeCompositionCache struct {
	sync.Mutex
	key     [sha256.Size]byte
	runtime []byte
}

func loadRuntimeComposition(resources fs.FS, templates map[string][]byte) ([]byte, error) {
	manifestPath := path.Join(runtimeCompositionDirectory, "support.json")
	raw, err := readAsset(resources, manifestPath)
	if err != nil {
		return nil, fmt.Errorf("load composed runtime manifest: %w", err)
	}
	var manifest runtimeComposition
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return nil, lerrors.InvalidConfiguration.Errorf("composed runtime manifest %s: %w", manifestPath, err)
	}
	if decoder.Decode(new(any)) != io.EOF {
		return nil, lerrors.InvalidConfiguration.Errorf("composed runtime manifest must contain one JSON value")
	}
	if manifest.Version != 1 || len(manifest.Sources) == 0 {
		return nil, lerrors.InvalidConfiguration.Errorf("composed runtime manifest requires version 1 and source fingerprints")
	}
	names := make([]string, 0, len(manifest.Sources))
	for name := range manifest.Sources {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !fs.ValidPath(name) || !strings.HasSuffix(name, ".l") || !(strings.HasPrefix(name, "builtins/") || strings.HasPrefix(name, "stdlib/")) {
			return nil, lerrors.InvalidConfiguration.Errorf("composed runtime source path %q must name a canonical source word file", name)
		}
		if _, err := readCompositionInput(resources, name, manifest.Sources[name]); err != nil {
			return nil, err
		}
	}
	aml, err := readCompositionInput(resources, path.Join(runtimeCompositionDirectory, "support.aml"), manifest.SHA256)
	if err != nil {
		return nil, err
	}
	system, err := loadSegmentalSystem(resources)
	if err != nil {
		return nil, fmt.Errorf("load composed runtime logemes: %w", err)
	}
	key := runtimeCompositionKey(raw, aml, system, templates)
	runtimeCompositionCache.Lock()
	defer runtimeCompositionCache.Unlock()
	if runtimeCompositionCache.runtime != nil && runtimeCompositionCache.key == key {
		return bytes.Clone(runtimeCompositionCache.runtime), nil
	}
	runtime, err := generateRuntimeComposition(aml, system, templates)
	if err != nil {
		return nil, err
	}
	if len(runtime) <= 1<<20 {
		runtimeCompositionCache.key = key
		runtimeCompositionCache.runtime = bytes.Clone(runtime)
	}
	return runtime, nil
}

func runtimeCompositionKey(manifest, aml []byte, system SegmentalSystem, templates map[string][]byte) [sha256.Size]byte {
	var fingerprints []byte
	add := func(content []byte) {
		digest := sha256.Sum256(content)
		fingerprints = append(fingerprints, digest[:]...)
	}
	add(manifest)
	add(aml)
	for _, definition := range segmentDefinitions {
		if definition.family == "Consonant" || definition.family == "Vowel" {
			add(definition.slot(&system).Content)
		}
	}
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		add([]byte(name))
		add(templates[name])
	}
	return sha256.Sum256(fingerprints)
}

func generateRuntimeComposition(aml []byte, system SegmentalSystem, templates map[string][]byte, names ...string) ([]byte, error) {
	scanner, err := loadScannerBytes(aml)
	if err != nil {
		return nil, fmt.Errorf("read composed runtime: %w", err)
	}
	if err := scanSourceProgram(&scanner); err != nil {
		return nil, fmt.Errorf("scan composed runtime: %w", err)
	}
	if len(scanner.Program.Books) != 1 || len(scanner.Program.Books[0].Outline.Services) != 1 {
		return nil, lerrors.InvalidConfiguration.Errorf("composed runtime requires one library target")
	}
	book := scanner.Program.Books[0]
	target := book.Outline.Services[0].Target
	if target == nil || target.Kind != "lib" || target.Root != "runtime" || target.Primary {
		return nil, lerrors.InvalidConfiguration.Errorf("composed runtime requires a sole library target rooted at runtime")
	}
	if len(names) == 1 && (book.Name != names[0] || target.Name != names[0]) {
		return nil, lerrors.InvalidConfiguration.Errorf("composed runtime must use its allocated package %s", names[0])
	}
	bundle, err := evaluateScannedSource(scanner, system, templates, false, nil)
	if err != nil {
		return nil, fmt.Errorf("generate composed runtime: %w", err)
	}
	if len(bundle.MapArrays) != 1 || len(bundle.ByteArrays) != 1 || len(bundle.MapArrays[0]) != len(bundle.ByteArrays[0]) {
		return nil, lerrors.InvalidConfiguration.Errorf("composed runtime requires one complete library bundle")
	}
	var runtime []byte
	for i, name := range bundle.MapArrays[0] {
		switch name {
		case "/src/runtime.rs":
			runtime = bundle.ByteArrays[0][i]
		case "/Cargo.toml":
			// Runtime dependencies remain the canonical host dependency set.
		default:
			return nil, lerrors.InvalidConfiguration.Errorf("composed runtime supplied unexpected support path %q", name)
		}
	}
	if len(runtime) == 0 {
		return nil, lerrors.InvalidConfiguration.Errorf("composed runtime requires generated /src/runtime.rs")
	}
	return append(bytes.Clone(runtime), []byte("\n// Native host adaptation is confined to the existing runtime module.\npub(crate) mod host;\n")...), nil
}

func readCompositionInput(resources fs.FS, name, expected string) ([]byte, error) {
	digest, err := hex.DecodeString(expected)
	if err != nil || len(digest) != sha256.Size {
		return nil, lerrors.InvalidConfiguration.Errorf("composed runtime input %s requires a SHA-256 fingerprint", name)
	}
	raw, err := readAsset(resources, name)
	if err != nil {
		return nil, fmt.Errorf("load composed runtime input %s: %w", name, err)
	}
	actual := sha256.Sum256(raw)
	if !bytes.Equal(actual[:], digest) {
		return nil, lerrors.InvalidConfiguration.Errorf("composed runtime input %s is stale; regenerate its canonical AML", name)
	}
	return raw, nil
}
