package dnslogger

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/replacer"
)

func init() { plugin.Register("dnslogger", setup) }

func setup(c *caddy.Controller) error {
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return DNSLogger{Next: next, repl: replacer.New()}
	})

	return nil
}
