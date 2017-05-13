package dns

import (
	"os"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

// Writer writes dns packetc in PCAP fromat.
type Writer struct {
	w *pcapgo.Writer
	f *os.File
}

// NewWriter creates a new Writer for dns packets.
func NewWriter(file string) (*Writer, error) {
	stat, err := os.Stat(file)
	return newWriter(file, os.IsNotExist(err) || stat.Size() == 0)
}

func newWriter(file string, writeHeader bool) (*Writer, error) {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	w := pcapgo.NewWriter(f)
	if writeHeader {
		// set file header only for new files
		if err = w.WriteFileHeader(65536, layers.LinkTypeEthernet); err != nil {
			return nil, err
		}
	}
	return &Writer{w, f}, nil
}

// Write writes slice of packets.
func (w *Writer) Write(packets []*Packet) error {
	if w == nil {
		return nil
	}

	for _, packet := range packets {
		if err := w.w.WritePacket(packet.raw.Metadata().CaptureInfo, packet.raw.Data()); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the file for saving packets.
func (w *Writer) Close() error {
	return w.f.Close()
}
