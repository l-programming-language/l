package Logiceme

import (
	"bytes"
	"testing"
)

func TestNativeMaskRetainsLiteralAndCommentBoundaries(t *testing.T) {
	source := []byte("/*| ignored : ~opening |*\\ E{|Statement: ~raw </W|> /*| and : ~ plus ~ interior\r\n雪~|}")
	before := bytes.Clone(source)
	masked, restore, err := maskNativeValues(source)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte(restore.Replace(string(masked))), source) || !bytes.Equal(source, before) {
		t.Fatal("native masking changed original literal or comment bytes")
	}
	if bytes.Contains(masked, []byte("</W|>")) || !bytes.Contains(masked, []byte("/*| ignored : ~opening |*\\")) {
		t.Fatalf("mask confused literal and structural comment boundaries: %q", masked)
	}
	word := append(append([]byte("<|W|>"), source...), []byte("</W|>tail")...)
	end, err := framedWordEnd(word, 0)
	if err != nil || string(word[end:]) != "tail" {
		t.Fatalf("word framing consumed native marker or tail: end=%d, %v", end, err)
	}
}
