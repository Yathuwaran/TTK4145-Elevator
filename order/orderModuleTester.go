package main

import (
	"TTK4145-Elevator/order/orderModule"
	S "TTK4145-Elevator/structs"
	"fmt"
)

func main() {

	//Dummy struct
	var elevOne S.Message_struct
	elevOne.ID = "ElevOne"
	elevOne.Last_floor = 4
	elevOne.Dir = S.MD_Down
	elevOne.State = S.Moving

	var elevTwo S.Message_struct
	elevTwo.ID = "ElevTwo"
	elevTwo.Last_floor = 1
	elevTwo.Dir = S.MD_Up
	elevTwo.State = S.Moving

	var elevThree S.Message_struct
	elevThree.ID = "ElevThree"
	elevThree.Last_floor = 2
	elevThree.Dir = S.MD_Up
	elevThree.State = S.Moving

	var order S.Order
	order.Button = S.BT_HallDown
	order.Floor = 3

	elevArray := []S.Message_struct{elevOne, elevTwo, elevThree}
	bestElev := orderModule.GetBestElev(elevArray, order)

	fmt.Println(bestElev.ID)
}
