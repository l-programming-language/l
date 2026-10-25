package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
)

type byteDifference struct {
	Equal           bool   `json:"equal"`
	ExpectedPath    string `json:"expected_path"`
	ActualPath      string `json:"actual_path"`
	ExpectedLength  int    `json:"expected_length"`
	ActualLength    int    `json:"actual_length"`
	ExpectedSHA256  string `json:"expected_sha256"`
	ActualSHA256    string `json:"actual_sha256"`
	FirstDifference *int   `json:"first_difference_offset,omitempty"`
	ExpectedContext string `json:"expected_context,omitempty"`
	ActualContext   string `json:"actual_context,omitempty"`
}

func compareBytes(expected, actual []byte) byteDifference {
	expectedSum, actualSum := sha256.Sum256(expected), sha256.Sum256(actual)
	r := byteDifference{
		Equal:          bytes.Equal(expected, actual),
		ExpectedLength: len(expected), ActualLength: len(actual),
		ExpectedSHA256: hex.EncodeToString(expectedSum[:]), ActualSHA256: hex.EncodeToString(actualSum[:]),
	}
	if r.Equal {
		return r
	}
	i := 0
	for i < len(expected) && i < len(actual) && expected[i] == actual[i] {
		i++
	}
	r.FirstDifference = &i
	r.ExpectedContext, r.ActualContext = byteContext(expected, i), byteContext(actual, i)
	return r
}

func byteContext(data []byte, position int) string {
	start, end := position-24, position+25
	if start < 0 {
		start = 0
	}
	if start > len(data) {
		start = len(data)
	}
	if end > len(data) {
		end = len(data)
	}
	return strconv.QuoteToASCII(string(data[start:end]))
}

func compareFiles(expectedPath, actualPath string) (byteDifference, error) {
	expected, err := os.ReadFile(expectedPath)
	if err != nil {
		return byteDifference{}, err
	}
	actual, err := os.ReadFile(actualPath)
	if err != nil {
		return byteDifference{}, err
	}
	r := compareBytes(expected, actual)
	r.ExpectedPath, r.ActualPath = expectedPath, actualPath
	return r, nil
}
