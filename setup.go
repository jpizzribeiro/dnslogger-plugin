package dnslogger

import (
	"fmt"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"strings"
)

func init() {
	caddy.RegisterPlugin("dnslogger", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
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

	// Configurar inicialização e finalização
	c.OnStartup(func() error {
		fmt.Printf("DNSLogger initialized with socket address: %s\n", socketAddr)
		return nil
	})

	c.OnShutdown(func() error {
		return client.Close()
	})

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return DNSLogger{Next: next, Client: client, SocketAddr: socketAddr}
	})

	// All OK, return a nil error.
	return nil
}
