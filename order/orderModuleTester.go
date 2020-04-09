package main

import (
	"./orderModule"
	S "../structs"
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

	type elevsd map[string]*S.Message_struct

	elevArray := make(elevsd)



	elevArray["ElevOne"] = &elevOne
	elevArray["ElevTwo"] = &elevTwo
	elevArray["ElevThree"] = &elevThree

	bestElev := orderModule.GetBestElev(elevArray, order)

	fmt.Println(bestElev.ID)
}
