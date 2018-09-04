package goast_test

import (
	"testing"

	"github.com/flowdev/go2md/goast"
)

func TestExtractFlowDSL(t *testing.T) {
	specs := []struct {
		name          string
		givenDoc      string
		expectedStart string
		expectedFlow  string
		expectedEnd   string
	}{
		{
			name:          "empty",
			givenDoc:      "",
			expectedStart: "",
			expectedFlow:  "",
			expectedEnd:   "",
		},
	}
	for _, spec := range specs {
		t.Logf("Testing doc: %s\n", spec.name)
		gotStart, gotFlow, gotEnd := goast.ExtractFlowDSL(spec.givenDoc)

		if spec.expectedStart != gotStart {
			t.Errorf("Expected start '%s', got '%s'.", spec.expectedStart, gotStart)
		}
		if spec.expectedFlow != gotFlow {
			t.Errorf("Expected flow '%s', got '%s'.", spec.expectedFlow, gotFlow)
		}
		if spec.expectedEnd != gotEnd {
			t.Errorf("Expected end '%s', got '%s'.", spec.expectedEnd, gotEnd)
		}
	}
}
