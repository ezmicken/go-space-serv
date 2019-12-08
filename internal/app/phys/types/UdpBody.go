package phys

type UdpBody struct {
	id int						// 4
	xPos float64			// 8
	yPos float64			// 8
	xVel float64			// 8
	yVel float64			// 8
	xAcc float64			// 8 - drop this if necessary
	yAcc float64			// 8 - drop this if necessary
}										// 50
