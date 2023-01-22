package feature

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmailFeatureActiveFor(t *testing.T) {
	tests := []struct {
		name  string
		cfg   EmailFeatureConfig
		email string
		want  bool
	}{
		{
			name:  "by email",
			cfg:   EmailFeatureConfig{Emails: []string{"sometest@mail.ru"}},
			email: "sometest@mail.ru",
			want:  true,
		}, {
			name:  "by domain",
			cfg:   EmailFeatureConfig{Domains: []string{"mail.ru"}},
			email: "sometest@mail.ru",
			want:  true,
		}, {
			name:  "by regexp",
			cfg:   EmailFeatureConfig{Regexps: []string{"^sometest.*"}},
			email: "sometest@mail.ru",
			want:  true,
		}, {
			name:  "by permille",
			cfg:   EmailFeatureConfig{Permille: 1000},
			email: "sometest@mail.ru",
			want:  true,
		}, {
			name: "active by email with conflict domain",
			cfg: EmailFeatureConfig{
				Emails:     []string{"sometest@mail.ru"},
				NotDomains: []string{"mail.ru"},
			},
			email: "sometest@mail.ru",
			want:  true,
		}, {
			name:  "inactive by not email",
			cfg:   EmailFeatureConfig{NotEmails: []string{"notsometest@corp.mail.ru"}},
			email: "notsometest@corp.mail.ru",
			want:  false,
		}, {
			name:  "inactive by not in domain",
			cfg:   EmailFeatureConfig{NotDomains: []string{"corp.mail.ru"}},
			email: "notsometest@corp.mail.ru",
			want:  false,
		}, {
			name: "inactive by not email with conflict domain",
			cfg: EmailFeatureConfig{
				NotEmails: []string{"notsometest@corp.mail.ru"},
				Domains:   []string{"corp.mail.ru"},
			},
			email: "notsometest@corp.mail.ru",
			want:  false,
		}, {
			name: "inactive by not email with suitable regexp",
			cfg: EmailFeatureConfig{
				NotEmails: []string{"notsometest@corp.mail.ru"},
				Regexps:   []string{".*sometest.*"},
			},
			email: "notsometest@corp.mail.ru",
			want:  false,
		}, {
			name: "inactive by not email with conflict email",
			cfg: EmailFeatureConfig{
				NotEmails: []string{"notsometest@corp.mail.ru"},
				Emails:    []string{"notsometest@corp.mail.ru"},
			},
			email: "notsometest@corp.mail.ru",
			want:  false,
		}, {
			name: "inactive by not email with conflict permille",
			cfg: EmailFeatureConfig{
				NotEmails: []string{"notsometest@corp.mail.ru"},
				Permille:  1000,
			},
			email: "notsometest@corp.mail.ru",
			want:  false,
		}, {
			name: "inactive by not domain with conflict domain",
			cfg: EmailFeatureConfig{
				NotDomains: []string{"corp.mail.ru"},
				Domains:    []string{"mail.ru"},
			},
			email: "notsometest@corp.mail.ru",
			want:  false,
		}, {
			name: "inactive by not domain with suitable regexp",
			cfg: EmailFeatureConfig{
				NotDomains: []string{"mail.ru"},
				Regexps:    []string{"^sometest.*"},
			},
			email: "sometest@mail.ru",
			want:  false,
		}, {
			name: "inactive by not domain with permille",
			cfg: EmailFeatureConfig{
				NotDomains: []string{"corp.mail.ru"},
				Permille:   1000,
			},
			email: "notsometest@corp.mail.ru",
			want:  false,
		}, {
			name:  "inactive",
			cfg:   EmailFeatureConfig{},
			email: "sometest@mail.ru",
			want:  false,
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(tt *testing.T) {
			f := NewEmailFeature(testCase.cfg)
			require.Equal(tt, f.ActiveFor(testCase.email), testCase.want)
		})
	}
}

// BenchmarkEmailFeatureActiveForRe
// BenchmarkEmailFeatureActiveForRe-8   	 1772512	       623 ns/op
func BenchmarkEmailFeatureActiveForRe(b *testing.B) {
	f := NewEmailFeature(EmailFeatureConfig{Regexps: []string{"^some.*@mail.ru$"}})

	for i := 0; i < b.N; i++ {
		f.ActiveFor("sometest@mail.ru")
	}
}

// BenchmarkEmailFeatureActiveForPercent
// BenchmarkEmailFeatureActiveForPercent-8   	 6187369	       170 ns/op
func BenchmarkEmailFeatureActiveForPercent(b *testing.B) {
	f := NewEmailFeature(EmailFeatureConfig{Permille: 100})

	for i := 0; i < b.N; i++ {
		f.ActiveFor("sometest@mail.ru")
	}
}
