package dnslogger

import (
	"fmt"
	"github.com/coredns/coredns/core/dnsserver"
	"strings"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
)

func init() {
	plugin.Register("dnslogger", setup)
}

func setup(c *caddy.Controller) error {
	var socketAddr string

	// Parse argumentos do plugin
	for c.Next() {
		if c.NextArg() {
			socketAddr = c.Val()
		} else {
			return plugin.Error("dnslogger", fmt.Errorf("missing socket address"))
		}
	}

	if socketAddr == "" {
		return plugin.Error("dnslogger", fmt.Errorf("socket address is required"))
	}

	// Criar cliente UDP
	client, err := NewUDPClient(strings.TrimSpace(socketAddr))
	if err != nil {
		return plugin.Error("dnslogger", fmt.Errorf("failed to create UDP client: %v", err))
	}

	//dnsLogger := &DNSLogger{
	//	SocketAddr: socketAddr,
	//	Client:     client,
	//}

	// Configurar inicialização e finalização
	c.OnStartup(func() error {
		clog.Infof("DNSLogger initialized with socket address: %s\n", socketAddr)
		return nil
	})

	c.OnShutdown(func() error {
		return client.Close()
	})

	// Registrar o handler
	//c.AddPlugin(func(next plugin.Handler) plugin.Handler {
	//	dnsLogger.Next = next
	//	return dnsLogger
	//})
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return DNSLogger{Next: next, Client: client, SocketAddr: socketAddr}
	})

	return nil
}
