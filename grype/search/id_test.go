package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anchore/grype/grype/vulnerability"
)

func Test_ByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		input   vulnerability.Vulnerability
		wantErr require.ErrorAssertionFunc
		matches bool
		reason  string
	}{
		{
			name: "match",
			id:   "CVE-YEAR-1",
			input: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID: "CVE-YEAR-1",
				},
			},
			matches: true,
		},
		{
			name: "not match",
			id:   "CVE-YEAR-1",
			input: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID: "CVE-YEAR-2",
				},
			},
			matches: false,
			reason:  `vulnerability ID "CVE-YEAR-2" does not match expected ID "CVE-YEAR-1"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := ByID(tt.id)
			matches, reason, err := constraint.MatchesVulnerability(tt.input)
			wantErr := require.NoError
			if tt.wantErr != nil {
				wantErr = tt.wantErr
			}
			wantErr(t, err)
			assert.Equal(t, tt.matches, matches)
			assert.Equal(t, tt.reason, reason)
		})
	}
}
