package normalize_test

import (
	"testing"

	"github.com/navikt/nada-backend/pkg/normalize"
	"github.com/stretchr/testify/assert"
)

func TestSANameFromEmail(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "email is converted to corrected name format",
			input:    "nada@nav.no",
			expected: "nada-at-nav-no",
		},
		{
			name:     "name is truncated to 63 characters",
			input:    "nadanadanadanadanadanadanadanadanadanadanadanadanadanadanadanadanadanadanadanadanadanadanada",
			expected: "nadanadanadanadanadanadanadanadanadanadanadanadanadanadanadanad",
		},
		{
			name:     "dash and underscore are trimmed",
			input:    "-hello_",
			expected: "hello",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalize.Email(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}
