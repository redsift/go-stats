package router_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redsift/go-stats/router"
	"github.com/redsift/go-stats/router/rules"
	"github.com/redsift/go-stats/stats"
)

func TestRouter(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	c := stats.NewMockCollector(ctrl)
	c.EXPECT().Inform("test", "test")
	c.EXPECT().Inform("bla", "bla", "some:tag", "test:bla")

	c2 := stats.NewMockCollector(ctrl)
	c2.EXPECT().Inform("bla", "bla")
	c2.EXPECT().Inform("bla", "bla", "some:tag", "bla:test")

	rs := rules.Any(
		rules.ByName("test"),
		rules.ByTagName("test"),
	)

	r := router.New(
		router.To(c, rs),
		router.To(c2, rs.Not()),
	)

	r.Inform("test", "test")
	r.Inform("bla", "bla", "some:tag", "test:bla")
	r.Inform("bla", "bla")
	r.Inform("bla", "bla", "some:tag", "bla:test")
}
