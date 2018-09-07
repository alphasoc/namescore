package alerts

import (
	"net"
	"time"

	"github.com/alphasoc/nfr/client"
	"github.com/alphasoc/nfr/groups"
)

// AlertMapper maps response to internal alert struct.
type AlertMapper struct {
	groups *groups.Groups
}

// Alert represents alert api struct.
type Alert struct {
	Follow string  `json:"follow"`
	More   bool    `json:"more"`
	Events []Event `json:"events"`
}

// Event from alert.
type Event struct {
	Type      string            `json:"type"`
	EventType string            `json:"eventType"`
	Flags     []string          `json:"flags"`
	Groups    []Group           `json:"groups"`
	Threats   map[string]Threat `json:"threats"`

	// fileds common for dns and ip event
	Timestamp time.Time `json:"ts"`
	SrcIP     net.IP    `json:"srcIp"`

	// ip event fileds
	SrcPort  int    `json:"srcPort"`
	DstIP    net.IP `json:"destIp"`
	DstPort  int    `json:"destPort"`
	Protocol string `json:"proto"`
	BytesIn  int    `json:"bytesIn"`
	BytesOut int    `json:"bytesOut"`
	Ja3      string `json:"ja3"`

	// dns event fields
	Query      string `json:"query"`
	RecordType string `json:"recordType"`
}

func (e *Event) topThreat() (string, Threat) {
	var (
		topID     string
		topThreat Threat
	)

	for tid, threat := range e.Threats {
		if threat.Severity > topThreat.Severity {
			topID = tid
			topThreat = threat
		}
	}

	return topID, topThreat
}

// Threat for event.
type Threat struct {
	Severity    int    `json:"severity"`
	Description string `json:"desc"`
	Policy      bool   `json:"policy"`
}

// Group describe group event belongs to.
type Group struct {
	Label       string `json:"label"`
	Description string `json:"desc"`
}

// NewAlertMapper creates new alert mapper.
func NewAlertMapper(groups *groups.Groups) *AlertMapper {
	return &AlertMapper{groups: groups}
}

// Map maps client response to alert.
func (m *AlertMapper) Map(resp *client.AlertsResponse) *Alert {
	var alert = &Alert{
		Follow: resp.Follow,
		More:   resp.More,
		Events: make([]Event, len(resp.Alerts)),
	}

	for i := range resp.Alerts {
		alert.Events[i] = Event{
			Type:      "alert",
			EventType: resp.Alerts[i].EventType,
			Flags:     resp.Alerts[i].Wisdom.Flags,
			Threats:   make(map[string]Threat),
		}

		for _, threat := range resp.Alerts[i].Threats {
			alert.Events[i].Threats[threat] = Threat{
				Severity:    resp.Threats[threat].Severity,
				Description: resp.Threats[threat].Title,
				Policy:      resp.Threats[threat].Policy,
			}
		}

		switch resp.Alerts[i].EventType {
		case "dns":
			alert.Events[i].Timestamp = resp.Alerts[i].DNSEvent.Timestamp
			alert.Events[i].SrcIP = resp.Alerts[i].DNSEvent.SrcIP
			alert.Events[i].Query = resp.Alerts[i].DNSEvent.Query
			alert.Events[i].RecordType = resp.Alerts[i].DNSEvent.QType
		case "ip":
			alert.Events[i].Timestamp = resp.Alerts[i].IPEvent.Timestamp
			alert.Events[i].SrcIP = resp.Alerts[i].IPEvent.SrcIP
			alert.Events[i].SrcPort = resp.Alerts[i].IPEvent.SrcPort
			alert.Events[i].DstIP = resp.Alerts[i].IPEvent.DstIP
			alert.Events[i].DstPort = resp.Alerts[i].IPEvent.DstPort
			alert.Events[i].Protocol = resp.Alerts[i].IPEvent.Protocol
			alert.Events[i].BytesIn = resp.Alerts[i].IPEvent.BytesIn
			alert.Events[i].BytesOut = resp.Alerts[i].IPEvent.BytesOut
			alert.Events[i].Ja3 = resp.Alerts[i].IPEvent.Ja3
		}

		for _, group := range m.groups.FindGroupsBySrcIP(alert.Events[i].SrcIP) {
			alert.Events[i].Groups = append(alert.Events[i].Groups, Group{
				Label:       group.Name,
				Description: group.Label,
			})
		}
	}
	return alert
}
