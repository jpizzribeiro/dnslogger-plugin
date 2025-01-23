package dnslogger

import (
	"database/sql"
	"fmt"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"strings"

	_ "github.com/marcboeker/go-duckdb"
)

func init() {
	caddy.RegisterPlugin("dnslogger", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	d, err := parse(c)
	if err != nil {
		return plugin.Error("dnslogger", err)
	}

	if d.SocketAddr == "" {
		return plugin.Error("dnslogger", fmt.Errorf("socket address is required"))
	}

	if d.DuckDbPath == "" {
		return plugin.Error("dnslogger", fmt.Errorf("DuckDBPath is required"))
	}

	db, err := sql.Open("duckdb", d.DuckDbPath)
	if err != nil {
		return plugin.Error("dnslogger", err)
	}
	defer db.Close()
	d.DB = db

	// Criar cliente UDP
	client, err := NewUDPClient(strings.TrimSpace(d.SocketAddr))
	if err != nil {
		return plugin.Error("dnslogger", fmt.Errorf("failed to create UDP client: %v", err))
	}

	// Configurar inicialização e finalização
	c.OnStartup(func() error {
		fmt.Printf("DNSLogger initialized with socket address: %s\n", d.SocketAddr)
		return nil
	})

	c.OnShutdown(func() error {
		return client.Close()
	})

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		d.Next = next
		return d
	})

	// All OK, return a nil error.
	return nil
}

func parse(c *caddy.Controller) (*DNSLogger, error) {
	var d = &DNSLogger{}

	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "socket":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return nil, c.ArgErr()
				}
				d.SocketAddr = args[0]
				break
			case "duckdbpath":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return nil, c.ArgErr()
				}
				d.DuckDbPath = args[0]
			}
		}
	}

	return d, nil
}
