package builtin

import (
	"github.com/pantheon-systems/search-secrets/pkg/app/config"
	. "github.com/pantheon-systems/search-secrets/pkg/search/rulebuild"
)

// Target definitions
func targetDefinitions() (result map[TargetName]*config.TargetConfig) {
	return map[TargetName]*config.TargetConfig{

		Passwords: {
			KeyPatterns: []string{
				`pass.?(?:word|wd|phrase)?`,
				`secret`,
				`secure`,
				`creds`,
				`credential`,
			},
			ExcludeKeyPatterns: []string{
				`policy\b`,
				`name\b`,
				`file\b`,
				`path\b`,
				`pass(?:es|ing)`,
			},
			ValChars:               AnyChars(),
			ValLenMin:              5,
			ValLenMax:              64,
			ValEntropyMin:          0,
			SkipFilePathLikeValues: true,
			SkipVariableLikeValues: true,
		},

		APIKeysAndTokens: {
			KeyPatterns: []string{
				`api.?key`,
				`token`,
				`secret`,
			},
			ExcludeKeyPatterns: []string{
				`\bform.?token\b`,
				`name\b`,
				`file\b`,
				`path\b`,
			},
			ValChars:               Base64PeriodDashUnderscoreChars(),
			ValLenMin:              32,
			ValLenMax:              64,
			ValEntropyMin:          2,
			SkipFilePathLikeValues: true,
			SkipVariableLikeValues: true,
		},
	}
}
