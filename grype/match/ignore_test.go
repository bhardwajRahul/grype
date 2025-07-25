package match

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/anchore/grype/grype/pkg"
	"github.com/anchore/grype/grype/vulnerability"
	"github.com/anchore/syft/syft/file"
	syftPkg "github.com/anchore/syft/syft/pkg"
)

var (
	allMatches = []Match{
		{
			Vulnerability: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID:        "CVE-123",
					Namespace: "debian-vulns",
				},
				Fix: vulnerability.Fix{
					State: vulnerability.FixStateFixed,
				},
			},
			Package: pkg.Package{
				ID:        pkg.ID(uuid.NewString()),
				Name:      "dive",
				Version:   "0.5.2",
				Type:      "deb",
				Locations: file.NewLocationSet(file.NewLocation("/path/that/has/dive")),
			},
		},
		{
			Vulnerability: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID:        "CVE-456",
					Namespace: "ruby-vulns",
				},
				Fix: vulnerability.Fix{
					State: vulnerability.FixStateNotFixed,
				},
				RelatedVulnerabilities: []vulnerability.Reference{
					{
						ID: "CVE-123",
					},
				},
			},
			Package: pkg.Package{
				ID:       pkg.ID(uuid.NewString()),
				Name:     "reach",
				Version:  "100.0.50",
				Language: syftPkg.Ruby,
				Type:     syftPkg.GemPkg,
				Locations: file.NewLocationSet(file.NewVirtualLocation("/real/path/with/reach",
					"/virtual/path/that/has/reach")),
			},
		},
		{
			Vulnerability: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID:        "CVE-457",
					Namespace: "ruby-vulns",
				},
				Fix: vulnerability.Fix{
					State: vulnerability.FixStateWontFix,
				},
			},
			Package: pkg.Package{
				ID:       pkg.ID(uuid.NewString()),
				Name:     "beach",
				Version:  "100.0.51",
				Language: syftPkg.Ruby,
				Type:     syftPkg.GemPkg,
				Locations: file.NewLocationSet(file.NewVirtualLocation("/real/path/with/beach",
					"/virtual/path/that/has/beach")),
			},
		},
		{
			Vulnerability: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID:        "CVE-458",
					Namespace: "ruby-vulns",
				},
				Fix: vulnerability.Fix{
					State: vulnerability.FixStateUnknown,
				},
			},
			Package: pkg.Package{
				ID:       pkg.ID(uuid.NewString()),
				Name:     "speach",
				Version:  "100.0.52",
				Language: syftPkg.Ruby,
				Type:     syftPkg.GemPkg,
				Locations: file.NewLocationSet(file.NewVirtualLocation("/real/path/with/speach",
					"/virtual/path/that/has/speach")),
			},
		},
	}

	// For testing the match-type rules
	matchTypesMatches = []Match{
		// Direct match, not like a normal kernel header match
		{
			Vulnerability: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID:        "CVE-1",
					Namespace: "fake-redhat-vulns",
				},
				Fix: vulnerability.Fix{
					State: vulnerability.FixStateUnknown,
				},
			},
			Package: pkg.Package{
				ID:      pkg.ID(uuid.NewString()),
				Name:    "kernel-headers1",
				Version: "5.1.0",
				Type:    syftPkg.RpmPkg,
				Upstreams: []pkg.UpstreamPackage{
					{Name: "kernel2"},
				},
			},
			Details: []Detail{
				{
					Type: ExactDirectMatch,
				},
			},
		},
		{
			Vulnerability: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID:        "CVE-2",
					Namespace: "fake-deb-vulns",
				},
				Fix: vulnerability.Fix{
					State: vulnerability.FixStateUnknown,
				},
			},
			Package: pkg.Package{
				ID:      pkg.ID(uuid.NewString()),
				Name:    "kernel-headers2",
				Version: "5.1.0",
				Type:    syftPkg.DebPkg,
				Upstreams: []pkg.UpstreamPackage{
					{Name: "kernel2"},
				},
			},
			Details: []Detail{
				{
					Type: ExactIndirectMatch,
				},
			},
		},
		{
			Vulnerability: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID:        "CVE-1",
					Namespace: "npm-vulns",
				},
				Fix: vulnerability.Fix{
					State: vulnerability.FixStateUnknown,
				},
			},
			Package: pkg.Package{
				ID:      pkg.ID(uuid.NewString()),
				Name:    "npm1",
				Version: "5.1.0",
				Type:    syftPkg.NpmPkg,
			},
			Details: []Detail{
				{
					Type: CPEMatch,
				},
			},
		},
	}

	// For testing the match-type and upstream ignore rules
	kernelHeadersMatches = []Match{
		// RPM-like match similar to what we see from RedHat
		{
			Vulnerability: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID:        "CVE-2",
					Namespace: "fake-redhat-vulns",
				},
				Fix: vulnerability.Fix{
					State: vulnerability.FixStateUnknown,
				},
			},
			Package: pkg.Package{
				ID:      pkg.ID(uuid.NewString()),
				Name:    "kernel-headers",
				Version: "5.1.0",
				Type:    syftPkg.RpmPkg,
				Upstreams: []pkg.UpstreamPackage{
					{Name: "kernel"},
				},
			},
			Details: []Detail{
				{
					Type: ExactIndirectMatch,
				},
			},
		},
		// debian-like match, showing the kernel header package name w/embedded version
		{
			Vulnerability: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID:        "CVE-2",
					Namespace: "fake-debian-vulns",
				},
				Fix: vulnerability.Fix{
					State: vulnerability.FixStateUnknown,
				},
			},
			Package: pkg.Package{
				ID:      pkg.ID(uuid.NewString()),
				Name:    "linux-headers-5.2.0",
				Version: "5.2.1",
				Type:    syftPkg.DebPkg,
				Upstreams: []pkg.UpstreamPackage{
					{Name: "linux"},
				},
			},
			Details: []Detail{
				{
					Type: ExactIndirectMatch,
				},
			},
		},
		// linux-like match, similar to what we see from debian\ubuntu
		{
			Vulnerability: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID:        "CVE-3",
					Namespace: "fake-linux-vulns",
				},
				Fix: vulnerability.Fix{
					State: vulnerability.FixStateUnknown,
				},
			},
			Package: pkg.Package{
				ID:      pkg.ID(uuid.NewString()),
				Name:    "linux-azure-headers-generic",
				Version: "5.2.1",
				Type:    syftPkg.DebPkg,
				Upstreams: []pkg.UpstreamPackage{
					{Name: "linux-azure"},
				},
			},
			Details: []Detail{
				{
					Type: ExactIndirectMatch,
				},
			},
		},
	}

	// For testing the match-type and upstream ignore rules
	packageTypeMatches = []Match{
		{
			Vulnerability: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID:        "CVE-2",
					Namespace: "fake-redhat-vulns",
				},
				Fix: vulnerability.Fix{
					State: vulnerability.FixStateUnknown,
				},
			},
			Package: pkg.Package{
				ID:      pkg.ID(uuid.NewString()),
				Name:    "kernel-headers",
				Version: "5.1.0",
				Type:    syftPkg.RpmPkg,
			},
		},
		{
			Vulnerability: vulnerability.Vulnerability{
				Reference: vulnerability.Reference{
					ID:        "CVE-2",
					Namespace: "fake-debian-vulns",
				},
				Fix: vulnerability.Fix{
					State: vulnerability.FixStateUnknown,
				},
			},
			Package: pkg.Package{
				ID:      pkg.ID(uuid.NewString()),
				Name:    "linux-headers-5.2.0",
				Version: "5.2.1",
				Type:    syftPkg.DebPkg,
			},
		},
	}
)

func TestApplyIgnoreRules(t *testing.T) {
	cases := []struct {
		name                     string
		allMatches               []Match
		ignoreRules              []IgnoreRule
		expectedRemainingMatches []Match
		expectedIgnoredMatches   []IgnoredMatch
	}{
		{
			name:                     "no ignore rules",
			allMatches:               allMatches,
			ignoreRules:              nil,
			expectedRemainingMatches: allMatches,
			expectedIgnoredMatches:   nil,
		},
		{
			name:       "no applicable ignore rules",
			allMatches: allMatches,
			ignoreRules: []IgnoreRule{
				{
					Vulnerability: "CVE-789",
				},
				{
					Package: IgnoreRulePackage{
						Name:    "bashful",
						Version: "5",
						Type:    "npm",
					},
				},
				{
					Package: IgnoreRulePackage{
						Name:    "reach",
						Version: "3000",
					},
				},
			},
			expectedRemainingMatches: allMatches,
			expectedIgnoredMatches:   nil,
		},
		{
			name:       "ignore all matches",
			allMatches: allMatches,
			ignoreRules: []IgnoreRule{
				{
					Vulnerability: "CVE-123",
				},
				{
					Package: IgnoreRulePackage{
						Location: "/virtual/path/that/has/reach",
					},
				},
			},
			expectedRemainingMatches: []Match{
				allMatches[2],
				allMatches[3],
			},
			expectedIgnoredMatches: []IgnoredMatch{
				{
					Match: allMatches[0],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Vulnerability: "CVE-123",
						},
					},
				},
				{
					Match: allMatches[1],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Package: IgnoreRulePackage{
								Location: "/virtual/path/that/has/reach",
							},
						},
					},
				},
			},
		},
		{
			name:       "ignore related matches",
			allMatches: allMatches,
			ignoreRules: []IgnoreRule{
				{
					Vulnerability:  "CVE-123",
					IncludeAliases: true,
				},
			},
			expectedRemainingMatches: []Match{
				allMatches[2],
				allMatches[3],
			},
			expectedIgnoredMatches: []IgnoredMatch{
				{
					Match: allMatches[0],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Vulnerability:  "CVE-123",
							IncludeAliases: true,
						},
					},
				},
				{
					Match: allMatches[1],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Vulnerability:  "CVE-123",
							IncludeAliases: true,
						},
					},
				},
			},
		},
		{
			name:       "ignore subset of matches",
			allMatches: allMatches,
			ignoreRules: []IgnoreRule{
				{
					Vulnerability: "CVE-456",
				},
			},
			expectedRemainingMatches: []Match{
				allMatches[0],
				allMatches[2],
				allMatches[3],
			},
			expectedIgnoredMatches: []IgnoredMatch{
				{
					Match: allMatches[1],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Vulnerability: "CVE-456",
						},
					},
				},
			},
		},
		{
			name:       "ignore matches without fix",
			allMatches: allMatches,
			ignoreRules: []IgnoreRule{
				{FixState: string(vulnerability.FixStateNotFixed)},
				{FixState: string(vulnerability.FixStateWontFix)},
				{FixState: string(vulnerability.FixStateUnknown)},
			},
			expectedRemainingMatches: []Match{
				allMatches[0],
			},
			expectedIgnoredMatches: []IgnoredMatch{
				{
					Match: allMatches[1],
					AppliedIgnoreRules: []IgnoreRule{
						{
							FixState: "not-fixed",
						},
					},
				},
				{
					Match: allMatches[2],
					AppliedIgnoreRules: []IgnoreRule{
						{
							FixState: "wont-fix",
						},
					},
				},
				{
					Match: allMatches[3],
					AppliedIgnoreRules: []IgnoreRule{
						{
							FixState: "unknown",
						},
					},
				},
			},
		},
		{
			name:       "ignore matches on namespace",
			allMatches: allMatches,
			ignoreRules: []IgnoreRule{
				{Namespace: "ruby-vulns"},
			},
			expectedRemainingMatches: []Match{
				allMatches[0],
			},
			expectedIgnoredMatches: []IgnoredMatch{
				{
					Match: allMatches[1],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Namespace: "ruby-vulns",
						},
					},
				},
				{
					Match: allMatches[2],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Namespace: "ruby-vulns",
						},
					},
				},
				{
					Match: allMatches[3],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Namespace: "ruby-vulns",
						},
					},
				},
			},
		},
		{
			name:       "ignore matches on language",
			allMatches: allMatches,
			ignoreRules: []IgnoreRule{
				{
					Package: IgnoreRulePackage{
						Language: string(syftPkg.Ruby),
					},
				},
			},
			expectedRemainingMatches: []Match{
				allMatches[0],
			},
			expectedIgnoredMatches: []IgnoredMatch{
				{
					Match: allMatches[1],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Package: IgnoreRulePackage{
								Language: string(syftPkg.Ruby),
							},
						},
					},
				},
				{
					Match: allMatches[2],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Package: IgnoreRulePackage{
								Language: string(syftPkg.Ruby),
							},
						},
					},
				},
				{
					Match: allMatches[3],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Package: IgnoreRulePackage{
								Language: string(syftPkg.Ruby),
							},
						},
					},
				},
			},
		},
		{
			name:       "ignore matches on indirect match-type",
			allMatches: matchTypesMatches,
			ignoreRules: []IgnoreRule{
				{
					MatchType: ExactIndirectMatch,
				},
			},
			expectedRemainingMatches: []Match{
				matchTypesMatches[0], matchTypesMatches[2],
			},
			expectedIgnoredMatches: []IgnoredMatch{
				{
					Match: matchTypesMatches[1],
					AppliedIgnoreRules: []IgnoreRule{
						{
							MatchType: ExactIndirectMatch,
						},
					},
				},
			},
		},
		{
			name:       "ignore matches on cpe match-type",
			allMatches: matchTypesMatches,
			ignoreRules: []IgnoreRule{
				{
					MatchType: CPEMatch,
				},
			},
			expectedRemainingMatches: []Match{
				matchTypesMatches[0], matchTypesMatches[1],
			},
			expectedIgnoredMatches: []IgnoredMatch{
				{
					Match: matchTypesMatches[2],
					AppliedIgnoreRules: []IgnoreRule{
						{
							MatchType: CPEMatch,
						},
					},
				},
			},
		},
		{
			name:       "ignore matches on upstream name",
			allMatches: kernelHeadersMatches,
			ignoreRules: []IgnoreRule{
				{
					Package: IgnoreRulePackage{
						UpstreamName: "kernel",
					},
				},
				{
					Package: IgnoreRulePackage{
						UpstreamName: "linux-.*",
					},
				},
			},
			expectedRemainingMatches: []Match{
				kernelHeadersMatches[1],
			},
			expectedIgnoredMatches: []IgnoredMatch{
				{
					Match: kernelHeadersMatches[0],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Package: IgnoreRulePackage{
								UpstreamName: "kernel",
							},
						},
					},
				},
				{
					Match: kernelHeadersMatches[2],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Package: IgnoreRulePackage{
								UpstreamName: "linux-.*",
							},
						},
					},
				},
			},
		},
		{
			name:       "ignore matches on package type",
			allMatches: packageTypeMatches,
			ignoreRules: []IgnoreRule{
				{
					Package: IgnoreRulePackage{
						Type: string(syftPkg.RpmPkg),
					},
				},
			},
			expectedRemainingMatches: []Match{
				packageTypeMatches[1],
			},
			expectedIgnoredMatches: []IgnoredMatch{
				{
					Match: packageTypeMatches[0],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Package: IgnoreRulePackage{
								Type: string(syftPkg.RpmPkg),
							},
						},
					},
				},
			},
		},
		{
			name:       "ignore matches rpms for kernel-headers with kernel upstream",
			allMatches: kernelHeadersMatches,
			ignoreRules: []IgnoreRule{
				{
					Package: IgnoreRulePackage{
						Name:         "kernel-headers",
						UpstreamName: "kernel",
						Type:         string(syftPkg.RpmPkg),
					},
					MatchType: ExactIndirectMatch,
				},
				{
					Package: IgnoreRulePackage{
						Name:         "linux-.*-headers-.*",
						UpstreamName: "linux.*",
						Type:         string(syftPkg.DebPkg),
					},
					MatchType: ExactIndirectMatch,
				},
			},
			expectedRemainingMatches: []Match{
				kernelHeadersMatches[1],
			},
			expectedIgnoredMatches: []IgnoredMatch{
				{
					Match: kernelHeadersMatches[0],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Package: IgnoreRulePackage{
								Name:         "kernel-headers",
								UpstreamName: "kernel",
								Type:         string(syftPkg.RpmPkg),
							},
							MatchType: ExactIndirectMatch,
						},
					},
				},
				{
					Match: kernelHeadersMatches[2],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Package: IgnoreRulePackage{
								Name:         "linux-.*-headers-.*",
								UpstreamName: "linux.*",
								Type:         string(syftPkg.DebPkg),
							},
							MatchType: ExactIndirectMatch,
						},
					},
				},
			},
		},
		{
			name:       "ignore on name regex",
			allMatches: kernelHeadersMatches,
			ignoreRules: []IgnoreRule{
				{
					Package: IgnoreRulePackage{
						Name: "kernel-headers.*",
					},
				},
			},
			expectedRemainingMatches: []Match{
				kernelHeadersMatches[1],
				kernelHeadersMatches[2],
			},
			expectedIgnoredMatches: []IgnoredMatch{
				{
					Match: kernelHeadersMatches[0],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Package: IgnoreRulePackage{
								Name: "kernel-headers.*",
							},
						},
					},
				},
			},
		},
		{
			name:       "ignore on name regex, no matches",
			allMatches: kernelHeadersMatches,
			ignoreRules: []IgnoreRule{
				{
					Package: IgnoreRulePackage{
						Name: "foo.*",
					},
				},
			},
			expectedRemainingMatches: kernelHeadersMatches,
			expectedIgnoredMatches:   nil,
		},
		{
			name:       "ignore on name regex, line termination verification",
			allMatches: kernelHeadersMatches,
			ignoreRules: []IgnoreRule{
				{
					Package: IgnoreRulePackage{
						Name: "^kernel-header$",
					},
				},
			},
			expectedRemainingMatches: kernelHeadersMatches,
			expectedIgnoredMatches:   nil,
		},
		{
			name:       "ignore on name regex, line termination test match",
			allMatches: kernelHeadersMatches,
			ignoreRules: []IgnoreRule{
				{
					Package: IgnoreRulePackage{
						Name: "^kernel-headers$",
					},
				},
			},
			expectedRemainingMatches: []Match{
				kernelHeadersMatches[1],
				kernelHeadersMatches[2],
			},
			expectedIgnoredMatches: []IgnoredMatch{
				{
					Match: kernelHeadersMatches[0],
					AppliedIgnoreRules: []IgnoreRule{
						{
							Package: IgnoreRulePackage{
								Name: "^kernel-headers$",
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			actualRemainingMatches, actualIgnoredMatches := ApplyIgnoreRules(sliceToMatches(testCase.allMatches), testCase.ignoreRules)

			assertMatchOrder(t, testCase.expectedRemainingMatches, actualRemainingMatches.Sorted())
			assertIgnoredMatchOrder(t, testCase.expectedIgnoredMatches, actualIgnoredMatches)

		})
	}
}

func sliceToMatches(s []Match) Matches {
	matches := NewMatches()
	matches.Add(s...)
	return matches
}

var (
	exampleMatch = Match{
		Vulnerability: vulnerability.Vulnerability{
			Reference: vulnerability.Reference{ID: "CVE-2000-1234"},
		},
		Package: pkg.Package{
			ID:      pkg.ID(uuid.NewString()),
			Name:    "a-pkg",
			Version: "1.0",
			Locations: file.NewLocationSet(
				file.NewLocation("/some/path"),
				file.NewVirtualLocation("/some/path", "/some/virtual/path"),
			),
			Type: "rpm",
		},
	}
)

func TestIsRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// simple strings that should NOT be detected as regex
		{
			name:     "simple string",
			input:    "hello",
			expected: false,
		},
		{
			name:     "alphanumeric with dashes",
			input:    "kernel-headers",
			expected: false,
		},
		{
			name:     "alphanumeric with underscores",
			input:    "my_package_name",
			expected: false,
		},
		{
			name:     "version numbers",
			input:    "1.2.3",
			expected: false, // dots are no longer considered regex metacharacters
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "spaces only",
			input:    "   ",
			expected: false,
		},
		{
			name:     "numbers only",
			input:    "12345",
			expected: false,
		},
		{
			name:     "letters and numbers",
			input:    "abc123",
			expected: false,
		},
		{
			name:     "with slashes",
			input:    "path/to/file",
			expected: false,
		},
		{
			name:     "with colons",
			input:    "namespace:package",
			expected: false,
		},
		{
			name:     "with at symbol",
			input:    "user@domain.com",
			expected: false, // dots are no longer considered regex metacharacters
		},

		// strings with regex metacharacters that SHOULD be detected as regex
		{
			name:     "caret at start",
			input:    "^start",
			expected: true,
		},
		{
			name:     "dollar at end",
			input:    "end$",
			expected: true,
		},
		{
			name:     "asterisk wildcard",
			input:    "test*",
			expected: true,
		},
		{
			name:     "plus quantifier",
			input:    "test+",
			expected: true,
		},
		{
			name:     "question mark",
			input:    "test?",
			expected: true,
		},
		{
			name:     "dot wildcard",
			input:    "test.",
			expected: false, // dots are no longer considered regex metacharacters
		},
		{
			name:     "square brackets",
			input:    "test[abc]",
			expected: true,
		},
		{
			name:     "parentheses grouping",
			input:    "(test)",
			expected: true,
		},
		{
			name:     "curly braces quantifier",
			input:    "test{1,3}",
			expected: true,
		},
		{
			name:     "pipe alternation",
			input:    "test|other",
			expected: true,
		},
		{
			name:     "backslash escape",
			input:    "test\\",
			expected: true,
		},
		{
			name:     "multiple metacharacters",
			input:    "^test.*$",
			expected: true,
		},
		{
			name:     "complex regex pattern",
			input:    "kernel-headers.*",
			expected: true,
		},
		{
			name:     "anchored regex",
			input:    "^kernel-headers$",
			expected: true,
		},
		{
			name:     "character class",
			input:    "test[0-9]",
			expected: true,
		},

		// escaped character classes
		{
			name:     "escaped digit",
			input:    "\\d",
			expected: true,
		},
		{
			name:     "escaped non-digit",
			input:    "\\D",
			expected: true,
		},
		{
			name:     "escaped word character",
			input:    "\\w",
			expected: true,
		},
		{
			name:     "escaped non-word character",
			input:    "\\W",
			expected: true,
		},
		{
			name:     "escaped whitespace",
			input:    "\\s",
			expected: true,
		},
		{
			name:     "escaped non-whitespace",
			input:    "\\S",
			expected: true,
		},
		{
			name:     "escaped newline",
			input:    "\\n",
			expected: true,
		},
		{
			name:     "escaped carriage return",
			input:    "\\r",
			expected: true,
		},
		{
			name:     "escaped tab",
			input:    "\\t",
			expected: true,
		},
		{
			name:     "escaped form feed",
			input:    "\\f",
			expected: true,
		},
		{
			name:     "escaped vertical tab",
			input:    "\\v",
			expected: true,
		},
		{
			name:     "escaped character classes in longer string",
			input:    "prefix\\dpostfix",
			expected: true,
		},
		{
			name:     "multiple escaped classes",
			input:    "\\w+\\s*\\d+",
			expected: true,
		},

		// edge cases
		{
			name:     "single backslash",
			input:    "\\",
			expected: true,
		},
		{
			name:     "single caret",
			input:    "^",
			expected: true,
		},
		{
			name:     "single dollar",
			input:    "$",
			expected: true,
		},
		{
			name:     "single dot",
			input:    ".",
			expected: false, // dots are no longer considered regex metacharacters
		},
		{
			name:     "backslash followed by regular character",
			input:    "\\a",
			expected: true, // backslash is still a metacharacter
		},
		{
			name:     "backslash at end",
			input:    "test\\",
			expected: true,
		},
		{
			name:     "mixed metacharacters and escaped classes",
			input:    "^\\w+\\.\\d{2,}$",
			expected: true,
		},
		{
			name:     "real world package patterns",
			input:    "linux-.*",
			expected: true,
		},
		{
			name:     "real world upstream patterns",
			input:    "linux.*",
			expected: true,
		},
		{
			name:     "real world header patterns",
			input:    "linux-.*-headers-.*",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLikelyARegex(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestShouldIgnore(t *testing.T) {
	cases := []struct {
		name     string
		match    Match
		rule     IgnoreRule
		expected bool
	}{
		{
			name:     "empty rule",
			match:    exampleMatch,
			rule:     IgnoreRule{},
			expected: false,
		},
		{
			name:  "rule applies via vulnerability ID",
			match: exampleMatch,
			rule: IgnoreRule{
				Vulnerability: exampleMatch.Vulnerability.ID,
			},
			expected: true,
		},
		{
			name:  "rule applies via package name",
			match: exampleMatch,
			rule: IgnoreRule{
				Package: IgnoreRulePackage{
					Name: exampleMatch.Package.Name,
				},
			},
			expected: true,
		},
		{
			name:  "rule applies via package version",
			match: exampleMatch,
			rule: IgnoreRule{
				Package: IgnoreRulePackage{
					Version: exampleMatch.Package.Version,
				},
			},
			expected: true,
		},
		{
			name:  "rule applies via package type",
			match: exampleMatch,
			rule: IgnoreRule{
				Package: IgnoreRulePackage{
					Type: string(exampleMatch.Package.Type),
				},
			},
			expected: true,
		},
		{
			name:  "rule applies via package location real path",
			match: exampleMatch,
			rule: IgnoreRule{
				Package: IgnoreRulePackage{
					Location: exampleMatch.Package.Locations.ToSlice()[0].RealPath,
				},
			},
			expected: true,
		},
		{
			name:  "rule applies via package location virtual path",
			match: exampleMatch,
			rule: IgnoreRule{
				Package: IgnoreRulePackage{
					Location: exampleMatch.Package.Locations.ToSlice()[1].AccessPath,
				},
			},
			expected: true,
		},
		{
			name:  "rule applies via package location glob",
			match: exampleMatch,
			rule: IgnoreRule{
				Package: IgnoreRulePackage{
					Location: "/some/**",
				},
			},
			expected: true,
		},
		{
			name:  "rule applies via multiple fields",
			match: exampleMatch,
			rule: IgnoreRule{
				Vulnerability: exampleMatch.Vulnerability.ID,
				Package: IgnoreRulePackage{
					Type: string(exampleMatch.Package.Type),
				},
			},
			expected: true,
		},
		{
			name:  "rule doesn't apply despite some fields matching",
			match: exampleMatch,
			rule: IgnoreRule{
				Vulnerability: exampleMatch.Vulnerability.ID,
				Package: IgnoreRulePackage{
					Name:    "not-the-right-package",
					Version: exampleMatch.Package.Version,
				},
			},
			expected: false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := len(testCase.rule.IgnoreMatch(testCase.match)) > 0
			assert.Equal(t, testCase.expected, actual)
		})
	}
}
