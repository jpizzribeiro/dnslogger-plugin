package dnslogger

import (
	"context"
	"fmt"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/request"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("dnslogger")

// DNSLogger is an example plugin to show how to write a plugin.
type DNSLogger struct {
	Next       plugin.Handler
	SocketAddr string
	Client     *UDPClient
}

// ServeDNS implements the plugin.Handler interface. This method gets called when example is used
// in a Server.
func (dl DNSLogger) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// Captura o estado da requisição
	state := request.Request{W: w, Req: r}
	name := state.Name()
	qType := dns.TypeToString[state.QType()]

	// Registrar log no servidor
	rrw := dnstest.NewRecorder(w)
	rc, err := plugin.NextOrFailure(state.Name(), dl.Next, ctx, rrw, r)
	if err != nil {
		clog.Warningf("Error processing DNS request: %v", err)
		return rc, err
	}

	// Preparar log para envio
	logEntry := fmt.Sprintf("Received query: %s Type: %s", name, qType)
	clog.Info(logEntry)

	// Enviar log via UDP
	if dl.Client != nil {
		if err := dl.Client.Send(logEntry); err != nil {
			clog.Warningf("Error sending log via UDP: %v", err)
		}
	}

	return rc, nil
}

// Name implements the Handler interface.
func (dl DNSLogger) Name() string { return "dnslogger" }
