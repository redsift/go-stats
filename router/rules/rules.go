package rules

import (
	"strings"

	"github.com/redsift/go-stats/router"
	"github.com/samber/lo"
)

func All(rules ...router.Rule) router.Rule {
	return ByFn(func(name string, tags []string) bool {
		for _, rule := range rules {
			if !rule.Match(name, tags) {
				return false
			}
		}
		return true
	})
}

func Any(rules ...router.Rule) router.Rule {
	return ByFn(func(name string, tags []string) bool {
		if len(rules) == 0 {
			return true
		}
		for _, rule := range rules {
			if rule.Match(name, tags) {
				return true
			}
		}
		return false
	})
}

func Not(r router.Rule) router.Rule {
	return &not{r}
}

func ByFn(fn func(name string, tags []string) bool) router.Rule {
	return &ruleFn{
		match: fn,
	}
}

func ByTag(tag string) router.Rule {
	return ByFn(
		func(name string, tags []string) bool {
			return lo.Contains(tags, tag)
		},
	)
}

func ByTagName(tagName string) router.Rule {
	tagName += ":"
	return ByFn(
		func(name string, tags []string) bool {
			return lo.ContainsBy(tags, func(tag string) bool {
				return strings.HasPrefix(tag, tagName)
			})
		},
	)
}

func ByName(nameFilter string) router.Rule {
	return ByFn(
		func(name string, tags []string) bool {
			return name == nameFilter
		},
	)
}

func ByNameFold(nameFilter string) router.Rule {
	return ByFn(
		func(name string, tags []string) bool {
			return strings.EqualFold(name, nameFilter)
		},
	)
}
