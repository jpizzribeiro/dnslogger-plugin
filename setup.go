package dnslogger

import (
	"database/sql"
	"fmt"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"strings"

	_ "github.com/marcboeker/go-duckdb"
	"github.com/maypok86/otter"
)

func init() {
	caddy.RegisterPlugin("dnslogger", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func populateCategories(db *sql.DB) map[int]Category {
	var categories = make(map[int]Category)
	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err == nil {
			categories[id] = Category{Name: name, ID: id}
		}
	}

	return categories
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

	cache, err := otter.MustBuilder[string, *DuckRow](10_000).
		CollectStats().
		Cost(func(key string, value *DuckRow) uint32 {
			return 1
		}).
		WithVariableTTL().
		Build()
	if err != nil {
		return plugin.Error("dnslogger", err)
	}
	d.Cache = cache

	db, err := sql.Open("duckdb", d.DuckDbPath)
	if err != nil {
		return plugin.Error("dnslogger", err)
	}
	// defer db.Close()

	d.DB = db

	d.Categories = populateCategories(db)
	var sources = make(map[string]SourceConfig)
	var block = make(map[int]struct{})
	block[3] = struct{}{}
	block[20] = struct{}{}

	sources["127.0.0.1"] = SourceConfig{
		BlockCategories: block,
	}
	d.Sources = sources

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
		db.Close()
		return client.Close()
	})

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		d.Next = next
		d.Client = client
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
			case "duckdbpath":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return nil, c.ArgErr()
				}
				d.DuckDbPath = args[0]
			case "blockip":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return nil, c.ArgErr()
				}
				d.BlockIp = args[0]
			}
		}
	}

	return d, nil
}
