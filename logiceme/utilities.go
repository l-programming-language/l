/*
Author: Connor Sanders
Version: 0.0.1
Released: 08/01/2020

-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
		Logiceme Evaluator V0.0.1
-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
*/

package Logiceme

import (
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/l-programming-language/l/lerrors"
)

// ReadSysFile reads a local file and returns content bytes
func ReadSysFile(fileName string) ([]byte, error) {
	content, err := os.ReadFile(fileName)
	return content, lerrors.IO.Wrap(err, "")
}

// readAsset reads from the one selected compiler distribution. Host AML imports
// deliberately continue using ReadSysFile instead.
func readAsset(resources fs.FS, name string) ([]byte, error) {
	if resources == nil {
		return nil, lerrors.InvalidConfiguration.Errorf("asset filesystem is required for %s", name)
	}
	content, err := fs.ReadFile(resources, name)
	return content, lerrors.IO.Wrap(err, "")
}

// GenerateUuid generates a string uuid
func GenerateUuid() string {
	curId, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return curId.String()
}

// checkIfStrInt
func checkIfStrInt(checkInt string) (bool, int) {
	if reInt, err := strconv.Atoi(checkInt); err == nil {
		return true, reInt
	}
	return false, 0
}

// StringifyArray converts an array of string into a readable string
func StringifyArray(strArray []string) string {
	reStr := ""
	for ind, st := range strArray {
		if ind == 0 {
			reStr = st
		} else {
			reStr = reStr + ", " + st
		}
	}
	return reStr
}

// StringifyStringArray converts a lucid array into a readable string set
func StringifyStringArray(strArray string) string {
	return strings.TrimSuffix(strings.TrimPrefix(strArray, "&["), "]&")
}

// TrimPrefixes removes all instances of a given prefix from a string dynamically
func TrimPrefixes(str string, prefix string) string {
	for _, letter := range str {
		if string(letter) == prefix {
			str = strings.TrimPrefix(str, prefix)
		} else {
			break
		}
	}
	return str
}
