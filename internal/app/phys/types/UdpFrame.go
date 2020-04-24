package phys

// TODO: optimize this for each client
// (only send the bodies that each client is interested in)

type UdpFrame struct {
  seq uint16
  frame []*UdpBody
}

func NewUdpFrame(s uint16) *UdpFrame {
  var f UdpFrame

  f.seq = s
  f.frame = []*UdpBody{}

  return &f
}

func (f *UdpFrame) AddUdpBody(b *UdpBody) {
  f.frame = append(f.frame, b)
}

func (f *UdpFrame) AddUdpBodies(b []*UdpBody) {
  f.frame = append(f.frame, b...)
}

func (f *UdpFrame) Len() int {
  return len(f.frame);
}

func (f *UdpFrame) Serialize() []byte {
  var data []byte

  for _, bod := range f.frame {
    data = append(data, bod.Serialize()...)
  }

  return data
}
