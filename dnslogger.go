package dnslogger

import (
	"context"
	"fmt"

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
func (dl *DNSLogger) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	q := r.Question[0]
	logEntry := fmt.Sprintf("Received query: %s %s %d", q.Name, dns.TypeToString[q.Qtype], q.Qclass)

	// Enviar log via UDP
	if err := dl.Client.Send(logEntry); err != nil {
		fmt.Printf("Error sending log: %v\n", err)
	}

	// Continuar com o próximo plugin na cadeia
	return dl.Next.ServeDNS(ctx, w, r)
}

// Name retorna o nome do plugin
func (dl *DNSLogger) Name() string {
	return "dnslogger"
}
