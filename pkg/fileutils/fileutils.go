package fileutils

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
)

// Matches returns true if relFilePath matches any of the patterns
// and isn't excluded by any of the subsequent patterns.
func exclusion(pattern string) bool {
	return pattern[0] == '!'
}

func empty(pattern string) bool {
	return pattern == ""
}

func Matches(relFilePath string, patterns []string) (bool, error) {
	matched := false

	for _, pattern := range patterns {
		pattern = filepath.Clean(pattern)
		negative := false
		if empty(pattern) {
			continue
		}

		if exclusion(pattern) {
			if len(pattern) == 1 {
				logrus.Errorf("Illegal exclusion pattern: %s", pattern)
				return false, errors.New("Illegal exclusion pattern: !")
			}
			negative = true
			pattern = pattern[1:]
		}

		match, err := filepath.Match(pattern, relFilePath)
		if err != nil {
			logrus.Errorf("Error matching: %s (pattern: %s)", relFilePath, pattern)
			return false, err
		}
		if match {
			if filepath.Clean(relFilePath) == "." {
				logrus.Errorf("Can't exclude whole path")
				return false, nil
			}
		} else if !match {
			// Check to see if the pattern matches one of our parent dirs
			path := filepath.Dir(relFilePath)
			pathDirs := strings.Split(path, "/")
			patternDirs := strings.Split(pattern, "/")
			if path != "." && len(patternDirs) <= len(pathDirs) {
				m, _ := filepath.Match(strings.Join(patternDirs, "/"),
					strings.Join(pathDirs[:len(patternDirs)], "/"))
				match = m
			}
		}

		if match {
			matched = !negative
		}
	}

	if matched {
		logrus.Debugf("Skipping excluded path: %s", relFilePath)
	}
	return matched, nil
}
