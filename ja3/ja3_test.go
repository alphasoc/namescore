package ja3

import (
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/stretchr/testify/require"
)

// testPacketPacket0 is the packet:
//   04:10:16.589867 IP 10.0.14.129.49206 > 185.174.175.14.443: Flags [P.], seq 2134588155:2134588313, ack 135605544, win 64240, length 158
//   	0x0000:  0002 1647 96ef 001a 9206 5c7b 0800 4500  ...G......\{..E.
//   	0x0010:  00c6 07f5 4000 8006 70ff 0a00 0e81 b9ae  ....@...p.......
//   	0x0020:  af0e c036 01bb 7f3b 3afb 0815 2d28 5018  ...6...;:...-(P.
//   	0x0030:  faf0 68f2 0000 1603 0300 9901 0000 9503  ..h.............
//   	0x0040:  035a f4fb 7795 5f4f fb01 23b7 4f0e a49b  .Z..w._O..#.O...
//   	0x0050:  26b8 f407 a99a 98d3 40a0 2516 be06 43b0  &.......@.%...C.
//   	0x0060:  b800 002a 003c 002f 003d 0035 0005 000a  ...*.<./.=.5....
//   	0x0070:  c027 c013 c014 c02b c023 c02c c024 c009  .'.....+.#.,.$..
//   	0x0080:  c00a 0040 0032 006a 0038 0013 0004 0100  ...@.2.j.8......
//   	0x0090:  0042 ff01 0001 0000 0000 1500 1300 0010  .B..............
//   	0x00a0:  726f 6277 6173 736f 7464 696e 742e 7275  robwassotdint.ru
//   	0x00b0:  000a 0006 0004 0017 0018 000b 0002 0100  ................
//   	0x00c0:  000d 0010 000e 0401 0501 0201 0403 0503  ................
//   	0x00d0:  0203 0202                                ....
var testTLSPacketPacket = []byte{
	0x00, 0x02, 0x16, 0x47, 0x96, 0xef, 0x00, 0x1a, 0x92, 0x06, 0x5c, 0x7b, 0x08, 0x00, 0x45, 0x00,
	0x00, 0xc6, 0x07, 0xf5, 0x40, 0x00, 0x80, 0x06, 0x70, 0xff, 0x0a, 0x00, 0x0e, 0x81, 0xb9, 0xae,
	0xaf, 0x0e, 0xc0, 0x36, 0x01, 0xbb, 0x7f, 0x3b, 0x3a, 0xfb, 0x08, 0x15, 0x2d, 0x28, 0x50, 0x18,
	0xfa, 0xf0, 0x68, 0xf2, 0x00, 0x00, 0x16, 0x03, 0x03, 0x00, 0x99, 0x01, 0x00, 0x00, 0x95, 0x03,
	0x03, 0x5a, 0xf4, 0xfb, 0x77, 0x95, 0x5f, 0x4f, 0xfb, 0x01, 0x23, 0xb7, 0x4f, 0x0e, 0xa4, 0x9b,
	0x26, 0xb8, 0xf4, 0x07, 0xa9, 0x9a, 0x98, 0xd3, 0x40, 0xa0, 0x25, 0x16, 0xbe, 0x06, 0x43, 0xb0,
	0xb8, 0x00, 0x00, 0x2a, 0x00, 0x3c, 0x00, 0x2f, 0x00, 0x3d, 0x00, 0x35, 0x00, 0x05, 0x00, 0x0a,
	0xc0, 0x27, 0xc0, 0x13, 0xc0, 0x14, 0xc0, 0x2b, 0xc0, 0x23, 0xc0, 0x2c, 0xc0, 0x24, 0xc0, 0x09,
	0xc0, 0x0a, 0x00, 0x40, 0x00, 0x32, 0x00, 0x6a, 0x00, 0x38, 0x00, 0x13, 0x00, 0x04, 0x01, 0x00,
	0x00, 0x42, 0xff, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x15, 0x00, 0x13, 0x00, 0x00, 0x10,
	0x72, 0x6f, 0x62, 0x77, 0x61, 0x73, 0x73, 0x6f, 0x74, 0x64, 0x69, 0x6e, 0x74, 0x2e, 0x72, 0x75,
	0x00, 0x0a, 0x00, 0x06, 0x00, 0x04, 0x00, 0x17, 0x00, 0x18, 0x00, 0x0b, 0x00, 0x02, 0x01, 0x00,
	0x00, 0x0d, 0x00, 0x10, 0x00, 0x0e, 0x04, 0x01, 0x05, 0x01, 0x02, 0x01, 0x04, 0x03, 0x05, 0x03,
	0x02, 0x03, 0x02, 0x02,
}

func TestConvert(t *testing.T) {
	p := gopacket.NewPacket(testTLSPacketPacket, layers.LinkTypeEthernet, gopacket.Default)
	if p.ErrorLayer() != nil {
		t.Error("failed to decode packet:", p.ErrorLayer().Error())
	}

	require.Equal(t, "4d7a28d6f2263ed61de88ca66eb011e3", Convert(p), "invalid ja3 hash")
}
