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
		}, {
			name:          "start-only",
			givenDoc:      "Let's start it.\n",
			expectedStart: "Let's start it.\n",
			expectedFlow:  "",
			expectedEnd:   "",
		}, {
			name:          "almost-flow",
			givenDoc:      "Start the over-\nflow:\nEnd.\n",
			expectedStart: "Start the over-\nflow:\nEnd.\n",
			expectedFlow:  "",
			expectedEnd:   "",
		}, {
			name:          "empty-flow",
			givenDoc:      "Start\n\nflow:\nThe end.",
			expectedStart: "Start\n",
			expectedFlow:  "",
			expectedEnd:   "The end.\n",
		}, {
			name:          "no-end",
			givenDoc:      "Start\n\nflow:\n    flow",
			expectedStart: "Start\n",
			expectedFlow:  "flow\n",
			expectedEnd:   "",
		}, {
			name:          "all-parts",
			givenDoc:      "Start\n\nflow:\n    my flow\nEnd\n",
			expectedStart: "Start\n",
			expectedFlow:  "my flow\n",
			expectedEnd:   "End\n",
		}, {
			name: "flow-with-empty-lines",
			givenDoc: "Start\n\nflow:\n    flow start\n          \t\r\n" +
				"    flow middle\n\n    flow end\nEnd\n",
			expectedStart: "Start\n",
			expectedFlow:  "flow start\n\nflow middle\n\nflow end\n",
			expectedEnd:   "End\n",
		}, {
			name: "long-end",
			givenDoc: "Start\n\nflow:\n    my flow\n" +
				"   The end\ndoesn't want to come.\nBut it\nhas to.\n",
			expectedStart: "Start\n",
			expectedFlow:  "my flow\n",
			expectedEnd:   "   The end\ndoesn't want to come.\nBut it\nhas to.\n",
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
