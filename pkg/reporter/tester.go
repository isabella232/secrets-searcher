package reporter

import (
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/finder/rule"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
)

type (
    Tester struct {
        rules                []rule.Rule
        whitelistPath        structures.RegexpSet
        whitelistSecretIDSet structures.Set
        builder              *Builder
        db                   *database.Database
        log                  *logrus.Logger
    }
)

func NewTester(rules []rule.Rule, whitelistPath structures.RegexpSet, whitelistSecretIDSet structures.Set, db *database.Database, log *logrus.Logger) *Tester {
    return &Tester{
        rules:                rules,
        whitelistPath:        whitelistPath,
        whitelistSecretIDSet: whitelistSecretIDSet,
        builder:              NewBuilder("app", db, log),
        db:                   db,
        log:                  log,
    }
}

func (t *Tester) Run() (err error) {
    // TODO: To allow testing of rule config while a search is running, pull data from current search and run it
    // TODO: against updated config
    return
}
