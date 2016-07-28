package exporter

import (
	"testing"
)

func TestAllRegexpsCompile(t *testing.T) {
	patterns := InitPatterns()
	loadPatternDir(t, patterns)
	for pattern := range *patterns {
		_, err := Compile("%{"+pattern+"}", patterns)
		if err != nil {
			t.Errorf("%v", err.Error())
		}
	}
}
