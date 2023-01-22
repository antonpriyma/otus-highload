package feature

import (
	"hash/crc32"
	"regexp"
	"strings"

	"github.com/antonpriyma/otus-highload/pkg/utils"
)

const (
	testLoginPrefix = "test_login"
	testBoxPrefix   = "test.box"
)

type Feature interface {
	ActiveFor(string) bool
}

type EmailFeatureConfig struct {
	Emails     []string `mapstructure:"emails"`
	NotEmails  []string `mapstructure:"not_emails"`
	Domains    []string `mapstructure:"domains"`
	NotDomains []string `mapstructure:"not_domains"`
	Regexps    []string `mapstructure:"regexps"`
	Permille   int      `mapstructure:"permille"`
	TestBoxes  bool     `mapstructure:"test_boxes"`
}

type emailFeature struct {
	Emails       map[string]bool
	NotEmails    map[string]bool
	Domains      map[string]bool
	NotInDomains map[string]bool
	Regexps      []*regexp.Regexp
	Permille     int
	TestBoxes    bool
}

func NewEmailFeature(cfg EmailFeatureConfig) Feature {
	compiledRegexps := make([]*regexp.Regexp, 0, len(cfg.Regexps))
	for _, re := range cfg.Regexps {
		compiledRegexps = append(compiledRegexps, regexp.MustCompile(re))
	}

	return emailFeature{
		Emails:       utils.StringSliceToSet(cfg.Emails),
		Domains:      utils.StringSliceToSet(cfg.Domains),
		Regexps:      compiledRegexps,
		Permille:     cfg.Permille,
		NotEmails:    utils.StringSliceToSet(cfg.NotEmails),
		NotInDomains: utils.StringSliceToSet(cfg.NotDomains),
		TestBoxes:    cfg.TestBoxes,
	}
}

func (f emailFeature) ActiveFor(email string) bool {
	email = strings.ToLower(email)
	if f.NotEmails[email] {
		return false
	}

	if f.Emails[email] {
		return true
	}

	_, domain := utils.Split2(email, "@")
	if f.NotInDomains[domain] {
		return false
	}

	if f.Domains[domain] {
		return true
	}

	if f.TestBoxes && inTestBoxes(email) {
		return true
	}

	if inPermille(email, f.Permille) {
		return true
	}

	for _, re := range f.Regexps {
		if re.MatchString(email) {
			return true
		}
	}

	return false
}

func inPermille(key string, permille int) bool {
	checksum := crc32.ChecksumIEEE([]byte(key))
	return int(checksum%1000) < permille
}

func inTestBoxes(email string) bool {
	return strings.HasPrefix(email, testLoginPrefix) || strings.HasPrefix(email, testBoxPrefix)
}
