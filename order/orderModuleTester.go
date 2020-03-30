package main

import (
	"TTK4145-Elevator/order/orderModule"
	S "TTK4145-Elevator/structs"
	"fmt"
)

func main() {

	//Dummy struct
	var elevatorOne S.Message_struct
	elevatorOne.Last_floor = 4
	elevatorOne.Dir = S.MD_Down
	elevatorOne.State = S.Moving

	var elevatorTwo S.Message_struct
	elevatorTwo.Last_floor = 3
	elevatorTwo.Dir = S.MD_Stop
	elevatorTwo.State = S.LOST_CONN

	var order S.Order
	order.Button = S.BT_HallDown
	order.Floor = 1

	elevatorOneScore := orderModule.ElevatorCostFunction(elevatorOne, order)
	elevatorTwoScore := orderModule.ElevatorCostFunction(elevatorTwo, order)

	fmt.Println("Elevator 1: ", elevatorOneScore, "Elevator 2: ", elevatorTwoScore)
}
