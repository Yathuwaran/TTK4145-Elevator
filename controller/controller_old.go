package main

import (
	"fmt"

	"../elevio"
)

type order_com struct {
    order chan int
}



//import "../order_module"
//test commit
func drive_elev(queue chan int, floors chan int) {
	current_floor := 2
	floor_goal := -1
	for {
		select {
		case a := <-queue:
			if current_floor < a {
				elevio.SetMotorDirection(elevio.MD_Up)
			} else if current_floor > a {
				elevio.SetMotorDirection(elevio.MD_Down)
			}
			floor_goal = a

		case a := <-floors:
			current_floor = a
			if current_floor == floor_goal {
				elevio.SetMotorDirection(elevio.MD_Stop)
				fmt.Println("Delivered")
			}
		}
	}
}

func notmain() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	elev_queue := make(chan int)
	floors := make(chan int)

	go drive_elev(elev_queue, floors)

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			elev_queue <- a.Floor

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}
			elevio.SetMotorDirection(d)
			floors <- a

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}

func poll_buttons() {

}
