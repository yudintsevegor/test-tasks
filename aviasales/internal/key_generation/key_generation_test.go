package key_generation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	tt := assert.New(t)

	cases := []struct {
		input          string
		expectedOutput string
	}{
		{
			input:          "aBaB",
			expectedOutput: "a2b2",
		},
		{
			input:          "aabB",
			expectedOutput: "a2b2",
		},
		{
			input:          "",
			expectedOutput: "",
		},
		{
			input:          "fffff",
			expectedOutput: "f5",
		},
		{
			input:          "abcdef",
			expectedOutput: "a1b1c1d1e1f1",
		},
		{
			input:          "aaBBccDd",
			expectedOutput: "a2b2c2d2",
		},
	}

	for i, c := range cases {
		out := New(c.input)
		tt.Equalf(c.expectedOutput, out, "[%d] input: %s", i, c.input)
	}
}
