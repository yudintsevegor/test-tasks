package uniq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	tt := assert.New(t)

	type expectedOutput struct {
		result []string
		isOk   bool
	}

	cases := []struct {
		caseName       string
		inputOriginal  []string
		inputWithUniq  []string
		expectedOutput expectedOutput
	}{
		{
			caseName:       "with one uniq",
			inputOriginal:  []string{"a", "b", "c"},
			inputWithUniq:  []string{"a", "b", "c", "d"},
			expectedOutput: expectedOutput{result: []string{"d"}, isOk: true},
		},
		{
			caseName:       "without uniq",
			inputOriginal:  []string{"a", "b", "c"},
			inputWithUniq:  []string{"a", "b", "c"},
			expectedOutput: expectedOutput{result: nil, isOk: false},
		},
		{
			caseName:       "with all uniq",
			inputOriginal:  []string{},
			inputWithUniq:  []string{"a", "b", "c"},
			expectedOutput: expectedOutput{result: []string{"a", "b", "c"}, isOk: true},
		},
	}

	for i, c := range cases {
		out, ok := Find(c.inputOriginal, c.inputWithUniq)
		tt.Equalf(c.expectedOutput.result, out, "[%d] case name: %s", i, c.caseName)
		tt.Equalf(c.expectedOutput.isOk, ok, "[%d] case name: %s", i, c.caseName)
	}
}
