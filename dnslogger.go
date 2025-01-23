package dnslogger

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/request"
	"github.com/elliotwutingfeng/go-fasttld"
	"github.com/maypok86/otter"
	"strings"
	"time"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("dnslogger")
var extractor, _ = fasttld.New(fasttld.SuffixListParams{})

// DNSLogger is an example plugin to show how to write a plugin.
type DNSLogger struct {
	Next       plugin.Handler
	Categories map[int]Category
	Sources    map[string]SourceConfig
	BlockIp    string
	DuckDbPath string
	SocketAddr string
	Client     *UDPClient
	DB         *sql.DB
	Cache      otter.CacheWithVariableTTL[string, *DuckRow]
}

type SourceConfig struct {
	BlockCategories     map[int]struct{}
	WhitelistCategories map[int]struct{}
}

type Category struct {
	ID   int
	Name string
}

type DuckRow struct {
	Domain     string
	CategoryId int
}

func (dl DNSLogger) searchDomainOnDuck(name string) *DuckRow {
	log.Info("Searching domain on Duck")
	row := &DuckRow{}

	if strings.HasSuffix(name, ".") {
		name = strings.TrimSuffix(name, ".")
	}

	err := dl.DB.QueryRow(`
   SELECT domain, category_id FROM domains WHERE domain = ? LIMIT 1
	`, name).Scan(&row.Domain, &row.CategoryId)
	if err != nil {
		log.Infof("%s not found on Duck", name)
		log.Error(err)
		return nil
	}

	return row
}

// ServeDNS implements the plugin.Handler interface. This method gets called when example is used
// in a Server.
func (dl DNSLogger) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// Captura o estado da requisição
	state := request.Request{W: w, Req: r}
	ip := state.IP()
	name := state.Name()
	qType := dns.TypeToString[state.QType()]

	var cacheKey = fmt.Sprintf("%s_%s", ip, name)

	var row *DuckRow
	onCache := false

	tld, err := extractor.Extract(fasttld.URLParams{URL: name})
	if err == nil {
		log.Infof("Domain: %s", tld.Domain)
		log.Infof("RegisteredDomain: %s", tld.RegisteredDomain)
		log.Infof("SubDomain: %s", tld.SubDomain)

		rowOnCache, ok := dl.Cache.Get(cacheKey)
		if ok {
			onCache = true
			row = rowOnCache
		}

		if !onCache {
			row = dl.searchDomainOnDuck(strings.TrimSpace(name))
			if row != nil {
				log.Infof("Domain: %s", row.Domain)
				log.Infof("CategoryId: %d", row.CategoryId)
				log.Infof("Category: %s", dl.Categories[row.CategoryId].Name)
			} else if tld.RegisteredDomain != "" && tld.RegisteredDomain != name {
				row = dl.searchDomainOnDuck(tld.RegisteredDomain)
				if row != nil {
					log.Infof("WildCard Domain: %s", row.Domain)
					log.Infof("WildCard CategoryId: %d", row.CategoryId)
					log.Infof("WildCard Category: %s", dl.Categories[row.CategoryId].Name)
				}
			}
		}

		sourceIp, exists := dl.Sources[ip]
		if exists {
			if row != nil {
				dl.Cache.Set(cacheKey, row, time.Minute)

				_, ok := sourceIp.BlockCategories[row.CategoryId]
				if ok {
					log.Errorf("Precisa bloquear")
					m := new(dns.Msg).
						SetRcode(r, dns.RcodeSuccess).
						SetEdns0(4096, true)

					for _, question := range r.Question {
						if question.Qtype == dns.TypeA {
							// Criar um novo registro A com o IP de destino
							rr, err := dns.NewRR(fmt.Sprintf("%s A %s", question.Name, dl.BlockIp))
							if err != nil {
								log.Errorf("Erro ao criar registro A: %v", err)
								continue
							}
							// Adicionar o registro à seção de resposta
							m.Answer = append(m.Answer, rr)
						}
					}

					logEntry := fmt.Sprintf("BLOCK Domain: %s", name)
					if dl.Client != nil {
						if err := dl.Client.Send(logEntry + "\n"); err != nil {
							clog.Warningf("Error sending log via UDP: %v", err)
						}
					}

					w.WriteMsg(m)
					return dns.RcodeSuccess, nil
				}
			}
		}
	}

	// Registrar log no servidor
	rrw := dnstest.NewRecorder(w)
	rc, err := plugin.NextOrFailure(dl.Name(), dl.Next, ctx, rrw, r)
	if err != nil {
		clog.Warningf("Error processing DNS request: %v", err)
		return rc, err
	}

	// Preparar log para envio
	logEntry := fmt.Sprintf("Received query: %s Type: %s", name, qType)
	clog.Debug(logEntry)

	// Enviar log via UDP
	/*if dl.Client != nil {
		if err := dl.Client.Send(logEntry + "\n"); err != nil {
			clog.Warningf("Error sending log via UDP: %v", err)
		}
	}*/

	return rc, nil
}

// Name implements the Handler interface.
func (dl DNSLogger) Name() string { return "dnslogger" }
