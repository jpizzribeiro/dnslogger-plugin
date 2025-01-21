package dnslogger

import (
	"context"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"

	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
)

// DNSLogger é a estrutura principal do plugin
type DNSLogger struct {
	Next       plugin.Handler
	SocketAddr string
	Client     *UDPClient
}

// ServeDNS processa as requisições DNS
func (dl DNSLogger) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	name := state.Name()

	rrw := dnstest.NewRecorder(w)
	rc, err := plugin.NextOrFailure(dl.Name(), dl.Next, ctx, rrw, r)
	if err != nil {
		clog.Info("plugin.NextOrFinish err:", err)
	}
	clog.Info(name, rrw.Msg.Question, rc)
	// logEntry := fmt.Sprintf("Received query: %s %s %d", q.Name, dns.TypeToString[q.Qtype], q.Qclass)

	// Enviar log via UDP
	// if err := dl.Client.Send(logEntry); err != nil {
	// fmt.Printf("Error sending log: %v\n", err)
	// }

	// Continuar com o próximo plugin na cadeia
	return plugin.NextOrFailure(dl.Name(), dl.Next, ctx, w, r)
}

// Name retorna o nome do plugin
func (dl DNSLogger) Name() string {
	return "dnslogger"
}
