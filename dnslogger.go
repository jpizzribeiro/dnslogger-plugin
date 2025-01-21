package dnslogger

import (
	"context"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/plugin/pkg/replacer"
	"github.com/coredns/coredns/plugin/pkg/response"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

// DNSLogger is a basic request logging plugin.
type DNSLogger struct {
	Next plugin.Handler

	repl replacer.Replacer
}

// ServeDNS implements the plugin.Handler interface.
func (dl DNSLogger) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	clog.Info(state.IP(), state.Name())
	return plugin.NextOrFailure(dl.Name(), dl.Next, ctx, w, r)
}

// Name implements the Handler interface.
func (dl DNSLogger) Name() string { return "dnslogger" }

// Rule configures the logging plugin.
type Rule struct {
	NameScope string
	Class     map[response.Class]struct{}
	Format    string
}
