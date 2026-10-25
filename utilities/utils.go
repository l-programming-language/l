package utilities

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gofrs/uuid"
	"github.com/l-programming-language/l/lerrors"
)

func exitGracefully(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

// Check checks for an error and gracefully exists the program if so
func Check(e error) {
	if e != nil {
		exitGracefully(e)
	}
}

// CheckIfValidFile checks if the file inputted exists
func CheckIfValidFile(filename string) (bool, error) {
	// Checking if entered file is CSV by using the filepath package from the standard library
	if fileExtension := filepath.Ext(filename); fileExtension != ".l" {
		return false, lerrors.InvalidInput.Errorf("file %s is not L", filename)
	}
	// Checking if filepath entered belongs to an existing file. We use the Stat method from the os package (standard library)
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false, lerrors.IO.Wrap(err, fmt.Sprintf("file %s does not exist", filename))
		}
		return false, lerrors.IO.Wrap(err, "")
	}
	return true, nil
}

// ReadFile checks if the file exists and reads the file contents if so
func ReadFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return []byte{}, lerrors.IO.Wrap(err, "")
	}
	defer file.Close() //nolint:errcheck // The read result is authoritative; this descriptor has no buffered writes.
	b, err := io.ReadAll(file)
	if err != nil {
		return []byte{}, lerrors.IO.Wrap(err, "")
	}
	return b, nil
}

// GenerateUuid returns a generated adhoc stringed UUID
func GenerateUuid() string {
	curId, err := uuid.NewV4()
	Check(err)
	return curId.String()
}
