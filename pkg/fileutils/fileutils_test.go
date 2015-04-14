package fileutils

import (
	"testing"
)

func TestWildcardMatches(t *testing.T) {
	match, _ := Matches("fileutils.go", []string{"*"})
	if match != true {
		t.Errorf("failed to get a wildcard match, got %v", match)
	}
}

func TestPatternMatches(t *testing.T) {
	match, _ := Matches("fileutils.go", []string{"*.go"})
	if match != true {
		t.Errorf("failed to get a match, got %v", match)
	}
}

func TestExclusionPatternMatchesPatternAfter(t *testing.T) {
	match, _ := Matches("fileutils.go", []string{"*.go", "!fileutils.go"})
	if match != false {
		t.Errorf("failed to get false match on exclusion pattern, got %v", match)
	}
}

func TestExclusionPatternMatchesPatternBefore(t *testing.T) {
	match, _ := Matches("fileutils.go", []string{"!fileutils.go", "*.go"})
	if match != true {
		t.Errorf("failed to get true match on exclusion pattern, got %v", match)
	}
}

func TestSingleExclamationError(t *testing.T) {
	_, err := Matches("fileutils.go", []string{"!"})
	if err == nil {
		t.Errorf("failed to get an error for a single exclamation point, got %v", err)
	}
}
