package rules

import (
	"github.com/redsift/go-stats/router"
)

type not struct {
	router.Rule
}

func (n not) Match(name string, tags []string) bool {
	return !n.Rule.Match(name, tags)
}

func (n not) Not() router.Rule {
	return n.Rule
}

type ruleFn struct {
	match func(name string, tags []string) bool
}

func (rfn *ruleFn) Match(name string, tags []string) bool {
	return rfn.match(name, tags)
}

func (rfn *ruleFn) Not() router.Rule {
	return &not{rfn}
}
