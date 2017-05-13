// Package executor execs main loop in namescore
package executor

import (
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/alphasoc/namescore/client"
	"github.com/alphasoc/namescore/config"
	"github.com/alphasoc/namescore/dns"
	"github.com/alphasoc/namescore/events"
	"github.com/alphasoc/namescore/groups"
)

type Executor struct {
	c   client.Client
	cfg *config.Config

	eventsPoller *events.Poller
	eventsWriter events.Writer

	groups *groups.Groups

	dnsWriter *dns.Writer
	sniffer   *dns.Sniffer
	buf       *dns.PacketBuffer

	mx sync.Mutex
}

func New(c client.Client, cfg *config.Config) (*Executor, error) {
	groups, err := createGroups(cfg)
	if err != nil {
		return nil, err
	}

	eventsWriter, err := events.NewJSONFileWriter(cfg.Events.File)
	if err != nil {
		return nil, err
	}

	eventsPoller := events.NewPoller(c, eventsWriter)
	if err = eventsPoller.SetFollowDataFile(cfg.Data.File); err != nil {
		return nil, err
	}

	return &Executor{
		c:            c,
		cfg:          cfg,
		eventsWriter: eventsWriter,
		eventsPoller: eventsPoller,
		groups:       groups,
		buf:          dns.NewPacketBuffer(),
	}, nil
}

func (e *Executor) Start() error {
	log.Infof("creating sniffer for %s interface, port %d, protocols %v",
		e.cfg.Network.Interface, e.cfg.Network.Port, e.cfg.Network.Protocols)
	sniffer, err := dns.NewLiveSniffer(e.cfg.Network.Interface, e.cfg.Network.Protocols, e.cfg.Network.Port)
	if err != nil {
		return err
	}
	e.sniffer = sniffer
	e.sniffer.SetGroups(e.groups)

	if e.cfg.Queries.Failed.File != "" {
		if e.dnsWriter, err = dns.NewWriter(e.cfg.Queries.Failed.File); err != nil {
			return err
		}
	}

	go e.installSignalHandler()
	go e.startEventPoller(e.cfg.Events.PollInterval, e.cfg.Events.File, e.cfg.Data.File)
	go e.startPacketSender(e.cfg.Queries.FlushInterval)

	return e.do()
}

func (e *Executor) StartOffline() error {
	log.Infof("creating offline sniffer for %s interface, port %d, protocols %v",
		e.cfg.Network.Interface, e.cfg.Network.Port, e.cfg.Network.Protocols)
	sniffer, err := dns.NewLiveSniffer(e.cfg.Network.Interface, e.cfg.Network.Protocols, e.cfg.Network.Port)
	if err != nil {
		return err
	}
	e.sniffer = sniffer
	e.sniffer.SetGroups(e.groups)

	if e.cfg.Queries.Failed.File != "" {
		if e.dnsWriter, err = dns.NewWriter(e.cfg.Queries.Failed.File); err != nil {
			return err
		}
	}

	go e.startPacketWriter(e.cfg.Queries.FlushInterval)

	return e.do()
}

func (e *Executor) Send(file string) error {
	log.Infof("creating sniffer for %s file", file)
	sniffer, err := dns.NewOfflineSniffer(file, e.cfg.Network.Protocols, e.cfg.Network.Port)
	if err != nil {
		return err
	}
	e.sniffer = sniffer
	e.sniffer.SetGroups(e.groups)

	return e.do()
}

func (e *Executor) startEventPoller(interval time.Duration, logFile, dataFile string) {
	// event poller will return error from api or
	// wrinting to disk. In both cases log the error
	// and try again in a moment.
	for {
		if err := e.eventsPoller.Do(interval); err != nil {
			log.Errorln(err)
		}
	}
}

func (e *Executor) startPacketSender(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		e.sendPackets()
	}
}

func (e *Executor) startPacketWriter(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		packets := e.buf.Packets()
		if len(packets) == 0 {
			continue
		} else if e.dnsWriter == nil {
			log.Infof("no queries failed file set, %d queries will be discarded", len(packets))
		} else if err := e.dnsWriter.Write(packets); err != nil {
			log.Warnln(err)
			continue
		} else {
			log.Infof("%d queries wrote to file", len(packets))
		}
	}
}

func (e *Executor) sendPackets() {
	e.mx.Lock()
	packets := e.buf.Packets()
	e.mx.Unlock()

	if len(packets) == 0 {
		return
	}

	log.Infof("sending %d dns queries to analyze", len(packets))
	resp, err := e.c.Queries(dnsPacketsToQueries(packets))
	if err != nil {
		log.Errorln(err)

		if e.dnsWriter != nil {
			// try to write packets to file. On success resset
			// buffer, else keep packets in buffer.
			if err := e.dnsWriter.Write(packets); err != nil {
				log.Warnln(err)
			} else {
				log.Infof("%d dns queries wrote to file", len(packets))
				return
			}

		}

		// write unsaved packets back to buffer
		e.mx.Lock()
		e.buf.Write(packets...)
		e.mx.Unlock()
		return
	}

	if resp.Received == resp.Accepted {
		log.Infof("%d dns queries were successfully send", resp.Accepted)
	} else {
		log.Infof("%d of %d dns queries were send - rejected reason %v",
			resp.Accepted, resp.Received, resp.Rejected)
	}
}

func (e *Executor) do() error {
	for packet := range e.sniffer.Packets() {
		e.mx.Lock()
		l := e.buf.Write(packet)
		e.mx.Unlock()
		if l < e.cfg.Queries.BufferSize {
			continue
		}

		// do not wait for sending packets
		go e.sendPackets()
	}

	// send what left in the buffer
	// and wait for other gorutines to finish
	// thanks to mutex lock in sendPackets
	e.sendPackets()
	return nil
}

func (e *Executor) installSignalHandler() {
	// Unless writer is set, then no handler is needed
	if e.dnsWriter == nil {
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		packets := e.buf.Packets()
		if err := e.dnsWriter.Write(packets); err != nil {
			log.Warnln(err)
		} else {
			log.Infof("%d queries wrote to file", len(packets))
		}
		os.Exit(1)
	}
}

func createGroups(cfg *config.Config) (*groups.Groups, error) {
	if len(cfg.WhiteListConfig.Groups) == 0 {
		return nil, nil
	}

	log.Infof("found %d whiltelist groups", len(cfg.WhiteListConfig.Groups))
	gs := groups.New()
	for name, group := range cfg.WhiteListConfig.Groups {
		g := &groups.Group{
			Name:     name,
			Includes: group.Networks,
			Excludes: group.Exclude.Networks,
			Domains:  group.Exclude.Domains,
		}
		if err := gs.Add(g); err != nil {
			return nil, err
		}
	}
	return gs, nil
}

func dnsPacketsToQueries(packets []*dns.Packet) *client.QueriesRequest {
	qr := client.NewQueriesRequest()
	for i := range packets {
		qr.AddQuery(packets[i].ToRequestQuery())
	}
	return qr
}
