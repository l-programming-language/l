package native

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/l-programming-language/l/internal/timing"
)

var evidenceSequence uint64
var evidenceIdentityOnce sync.Once
var evidenceIdentity string
var evidenceIdentityErr error

const evidenceRunEnvironment = "L_HARNESS_RUN_ID"
const cargoEvidenceEnvironment = "L_HARNESS_CARGO_EVIDENCE_DIR"

func evidenceRunID() (string, error) {
	if id := os.Getenv(evidenceRunEnvironment); id != "" {
		if id == "." || id == ".." || !regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`).MatchString(id) {
			return "", fmt.Errorf("invalid %s", evidenceRunEnvironment)
		}
		return id, nil
	}
	evidenceIdentityOnce.Do(func() {
		var nonce [12]byte
		_, evidenceIdentityErr = rand.Read(nonce[:])
		evidenceIdentity = fmt.Sprintf("standalone-%d-%x", os.Getpid(), nonce)
	})
	return evidenceIdentity, evidenceIdentityErr
}

func evidenceWriteOnce(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

// ReadPublishedGeneration decodes the sole manifest into the caller's schema.
func ReadPublishedGeneration(t testing.TB, binary string, destination any) string {
	t.Helper()
	manifests, err := filepath.Glob(filepath.Join(binary+".apps", "generations", "*", "manifest.json"))
	if err != nil || len(manifests) != 1 {
		t.Fatalf("expected one immutable generation: %v %v", manifests, err)
	}
	data, err := os.ReadFile(manifests[0])
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, destination); err != nil {
		t.Fatal(err)
	}
	return manifests[0]
}

// NativeArtifactHash validates the native executable before hashing its bytes.
func NativeArtifactHash(t testing.TB, path string) string {
	t.Helper()
	AssertNativeArtifact(t, path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// RecordBundle retains every returned artifact (including unexpected paths,
// dependencies and nil/empty distinctions). It is diagnostic output only and
// is never read as a source expectation or passed back to the compiler.
func RecordBundle(t testing.TB, phase string, bundle any) {
	t.Helper()
	root := os.Getenv("L_HARNESS_EVIDENCE_DIR")
	if root == "" {
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatal("cannot locate parity evidence root")
		}
		root = filepath.Join(filepath.Dir(file), "../../../.harness/native-evidence")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	safe := regexp.MustCompile(`[^a-zA-Z0-9_.-]+`)
	runID, err := evidenceRunID()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, runID, "bundles", safe.ReplaceAllString(filepath.Base(filepath.Dir(cwd))+"_"+filepath.Base(cwd)+"_"+t.Name(), "_"))
	if err = os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	var nonce [12]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("%s-%d-%d-%x.json", safe.ReplaceAllString(phase, "_"), os.Getpid(), atomic.AddUint64(&evidenceSequence, 1), nonce)
	path := filepath.Join(dir, name)
	data = append(data, '\n')
	if err = evidenceWriteOnce(path, data); err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	record := map[string]any{"RunID": runID, "Path": path, "SHA256": digest, "Bytes": len(data), "PID": os.Getpid(), "PPID": os.Getppid(), "Test": t.Name(), "Phase": phase, "RecordedAt": time.Now().UTC()}
	metadata, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = evidenceWriteOnce(path+".record.json", append(metadata, '\n')); err != nil {
		t.Fatal(err)
	}
	timing.Record("bundle.recorded", path, record)
	t.Logf("complete actual bundle: %s (sha256 %s; run %s)", path, digest, runID)
}
