package phys

type UdpFrame struct {
	seq byte
	size int
	frame []UdpBody
}

func (f *UdpFrame) New(s byte) {
	f.seq = s;
	f.frame = []UdpBody{}
}

func (f *UdpFrame) AddUdpBody(b UdpBody) {
	f.frame = append(f.frame, b)
}
