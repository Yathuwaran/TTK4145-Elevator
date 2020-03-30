package orderModule

import (
	S "TTK4145-Elevator/structs"
	"math"
)

var orderQueue chan S.Order
var initStatus bool

// QueueInit initializes the orderQueue
func QueueInit() {
	if !initStatus {
		initStatus = true
		orderQueue = make(chan S.Order, 30)
	}
}

// EnqueueLocal puts a new order into the orderQueue
func EnqueueLocal(order S.Order) {
	QueueInit()
	orderQueue <- order
}

// RetrieveLocalOrder returns first order in queue
func RetrieveLocalOrder() S.Order {
	order := <-orderQueue
	return order
}

/*
func DetermineBestElevator() S.Message_struct {

}*/

func ElevatorCostFunction(peer S.Message_struct, order S.Order) int {

	var totalScore int

	const ( // TODO: ASSIGN VALUES
		Weight_PickupOnPassing       = 1000
		Weight_DistanceToOrder       = 10
		Weight_ActivityStateRunning  = 1
		Weight_ActivityStateIdle     = 50
		Weight_ActivityStateOpenDoor = 45
	)

	// A high weight-score means that the evaluated elevator is well suited for the order

	///////////////////////////////////////////////////////////////////////////////

	// If going down, and new order is on the way to existing order
	if (peer.Dir == S.MD_Down) && (order.Floor < peer.Last_floor) && ((order.Button == S.BT_Cab) || (order.Button == S.BT_HallDown)) {

		totalScore += (1 * Weight_PickupOnPassing) // This should weigh a lot

	}

	///////////////////////////////////////////////////////////////////////////////

	// If going up, and new order is on the way to existing order
	if (peer.Dir == S.MD_Up) && (order.Floor > peer.Last_floor) && ((order.Button == S.BT_Cab) || (order.Button == S.BT_HallUp)) {

		totalScore += (1 * Weight_PickupOnPassing) // This should weigh a lot

	}

	///////////////////////////////////////////////////////////////////////////////

	floorDifference := int(math.Abs(float64(order.Floor - peer.Last_floor)))
	totalScore += (4 - floorDifference) * Weight_DistanceToOrder

	///////////////////////////////////////////////////////////////////////////////

	if peer.State == S.Moving {

		totalScore += (1 * Weight_ActivityStateRunning)

	} else if peer.State == S.IDLE {

		totalScore += (1 * Weight_ActivityStateIdle)

	} else if peer.State == S.LOST_CONN {
		totalScore -= 10000
	} else if peer.State == S.OPENDOOR {
		totalScore += (1 * Weight_ActivityStateOpenDoor)
	}

	///////////////////////////////////////////////////////////////////////////////

	return totalScore

}
