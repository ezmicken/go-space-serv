package phys

import(
	"github.com/panjf2000/gnet"
)

type Body struct {
	Bod UdpBody
	Owner gnet.Conn
	Accelerates bool
	Bounces int
}
