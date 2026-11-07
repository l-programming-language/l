package distribution

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"sort"

	"github.com/l-programming-language/l/internal/sourcepaths"
	"github.com/l-programming-language/l/lerrors"
)

// RuntimeComposition records derived AML and its source-owned selection ranges.
// Selection is an optional resource index, not an AML field or native code.
type RuntimeComposition struct {
	Version   int                 `json:"version"`
	SHA256    string              `json:"sha256"`
	Sources   map[string]string   `json:"sources"`
	Selection map[string][][2]int `json:"selection,omitempty"`
}

// Fingerprint binds the complete source inventory and selection to their AML.
// It detects stale bundles, not forged provenance; explicit source generation
// remains responsible for deriving the inventory and checking the composition.
func (m RuntimeComposition) Fingerprint(aml []byte) (string, error) {
	m.SHA256 = ""
	header, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	content := append([]byte("L runtime composition v1\x00"), header...)
	content = append(content, 0)
	content = append(content, aml...)
	return fmt.Sprintf("%x", sha256.Sum256(content)), nil
}

// IsEmbedded identifies this immutable distribution without comparing arbitrary
// fs.FS values, which may contain non-comparable maps or slices.
func IsEmbedded(files fs.FS) bool {
	embedded, ok := files.(embed.FS)
	return ok && embedded == resources
}

// LoadRuntimeComposition validates every declared source before returning AML.
// Mutable providers are reread on every call; neither errors nor parsed state
// are cached by this resource reader.
func LoadRuntimeComposition(files fs.FS) (RuntimeComposition, []byte, error) {
	const manifestPath = "runtime/generated/support.json"
	var manifest RuntimeComposition
	if files == nil {
		return manifest, nil, lerrors.InvalidConfiguration.Errorf("asset filesystem is required for %s", manifestPath)
	}
	raw, err := fs.ReadFile(files, manifestPath)
	if err != nil {
		return manifest, nil, fmt.Errorf("load composed runtime manifest: %w", lerrors.IO.Wrap(err, ""))
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return manifest, nil, lerrors.InvalidConfiguration.Errorf("composed runtime manifest %s: %w", manifestPath, err)
	}
	if decoder.Decode(new(any)) != io.EOF {
		return manifest, nil, lerrors.InvalidConfiguration.Errorf("composed runtime manifest must contain one JSON value")
	}
	if manifest.Version != 1 || len(manifest.Sources) == 0 {
		return manifest, nil, lerrors.InvalidConfiguration.Errorf("composed runtime manifest requires version 1 and source fingerprints")
	}
	if manifest.Sources["dictionaries.json"] == "" {
		return manifest, nil, lerrors.InvalidConfiguration.Errorf("composed runtime manifest requires the dictionaries.json fingerprint")
	}
	names := make([]string, 0, len(manifest.Sources))
	for name := range manifest.Sources {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if name != "dictionaries.json" && !sourcepaths.Canonical(name) {
			return manifest, nil, lerrors.InvalidConfiguration.Errorf("composed runtime source path %q must name a canonical source word file", name)
		}
		if _, err := readCompositionInput(files, name, manifest.Sources[name]); err != nil {
			return manifest, nil, err
		}
	}
	const amlPath = "runtime/generated/support.aml"
	aml, err := fs.ReadFile(files, amlPath)
	if err != nil {
		return manifest, nil, fmt.Errorf("load composed runtime input %s: %w", amlPath, lerrors.IO.Wrap(err, ""))
	}
	fingerprint, err := manifest.Fingerprint(aml)
	if err == nil {
		err = checkCompositionFingerprint(amlPath, manifest.SHA256, fingerprint)
	}
	if err != nil {
		aml = nil
	}
	return manifest, aml, err
}

func readCompositionInput(files fs.FS, name, expected string) ([]byte, error) {
	raw, err := fs.ReadFile(files, name)
	if err != nil {
		return nil, fmt.Errorf("load composed runtime input %s: %w", name, lerrors.IO.Wrap(err, ""))
	}
	return raw, checkCompositionFingerprint(name, expected, fmt.Sprintf("%x", sha256.Sum256(raw)))
}

func checkCompositionFingerprint(name, expected, actual string) error {
	digest, err := hex.DecodeString(expected)
	if err != nil || len(digest) != sha256.Size {
		return lerrors.InvalidConfiguration.Errorf("composed runtime input %s requires a SHA-256 fingerprint", name)
	}
	if hex.EncodeToString(digest) != actual {
		return lerrors.InvalidConfiguration.Errorf("composed runtime input %s is stale; regenerate its canonical AML", name)
	}
	return nil
}
