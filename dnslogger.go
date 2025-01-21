package dnslogger

import (
	"context"
	"fmt"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
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
	// Captura o estado da requisição
	state := request.Request{W: w, Req: r}
	name := state.Name()
	qType := dns.TypeToString[state.QType()]

	// Registrar log no servidor
	//rrw := dnstest.NewRecorder(w)
	//rc, err := plugin.NextOrFailure(dl.Name(), dl.Next, ctx, rrw, r)
	//if err != nil {
	//	clog.Warningf("Error processing DNS request: %v", err)
	//	return rc, err
	//}

	// Preparar log para envio
	logEntry := fmt.Sprintf("Received query: %s Type: %s", name, qType)
	clog.Info(logEntry)

	// Enviar log via UDP
	if dl.Client != nil {
		if err := dl.Client.Send(logEntry); err != nil {
			clog.Warningf("Error sending log via UDP: %v", err)
		}
	}

	return plugin.NextOrFailure(dl.Name(), dl.Next, ctx, w, r)
}

// Name retorna o nome do plugin
func (dl DNSLogger) Name() string {
	return "dnslogger"
}
