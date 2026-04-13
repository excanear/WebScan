package web

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestLoadSignatures(t *testing.T) {
	// create a temporary signatures file
	tmp, err := ioutil.TempFile("", "sigs-*.json")
	if err != nil {
		t.Fatalf("tempfile: %v", err)
	}
	defer os.Remove(tmp.Name())
	content := `[
            {"name":"TestSig","body_contains":["marker"]}
        ]`
	if _, err := tmp.WriteString(content); err != nil {
		t.Fatalf("write tmp: %v", err)
	}
	tmp.Close()

	// preserve original DB
	orig := signatureDB
	defer func() { signatureDB = orig }()

	if err := LoadSignatures(tmp.Name()); err != nil {
		t.Fatalf("LoadSignatures: %v", err)
	}

	// should match based on body content
	matches := MatchSignatures(http.Header{}, []byte("marker"))
	found := false
	for _, m := range matches {
		if m == "TestSig" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected TestSig in matches, got: %v", matches)
	}
}
