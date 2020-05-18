package search

import (
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
)

type CodeWhitelist struct {
	Res *manip.RegexpSet
	log logg.Logg
}

func NewCodeWhitelist(values []string, log logg.Logg) (result *CodeWhitelist) {
	res := manip.NewRegexpSetFromStringsMustCompile(values)
	return &CodeWhitelist{Res: res, log: log}
}

func (f *CodeWhitelist) IsSecretWhitelisted(input string, lineRange *manip.LineRange) (result bool) {
	for _, re := range f.Res.ReValues() {

		matches := re.FindAllStringSubmatchIndex(input, -1)

		log := f.log.WithField("re", re.String()).WithField("str", input)

		if matches == nil {
			log.Trace("secret does not match whitelist")
			continue
		}
		matchesLen := len(matches)

		for i, match := range matches {
			n := i + 1

			matchLineRange := manip.NewLineRange(match[0], match[1])

			// If there's a backreference, it's location should match the provided location
			if len(match) > 2 {
				backrefLineRange := manip.NewLineRange(match[2], match[3])
				if backrefLineRange.Equals(lineRange) {
					log.Tracef("backref matches for match %d of %d", n, matchesLen)
					return true
				}

				log.Tracef("backref doesn't match for match %d of %d", n, matchesLen)

				continue
			}

			// If no backreference but the match overlaps, return true
			if matchLineRange.Overlaps(lineRange) {
				log.Tracef("no backref for match %d of %d but it overlaps", n, matchesLen)

				return true
			}

			log.Tracef("no backref, no overlap for match %d of %d", n, matchesLen)
		}
	}

	return false
}
