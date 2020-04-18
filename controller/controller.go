package controller


import (
	"fmt"
	"../elevio"
	"../structs"
	"time"
)


type Event struct {
	buttons chan elevio.ButtonEvent
	floors chan int
	obstr chan bool
	stop chan bool
}


func Init_elev(port string, numFloors int) (Event, int) {
	elevio.Init("localhost:"+port, numFloors)

	var d elevio.MotorDirection = elevio.MD_Down
	floor := -1

	event := Event{
		buttons:    make(chan elevio.ButtonEvent),
		floors:     make(chan int, 1024),
		obstr:      make(chan bool),
		stop:       make(chan bool)}

	go elevio.PollButtons(event.buttons)
	go elevio.PollFloorSensor(event.floors)
	go elevio.PollObstructionSwitch(event.obstr)
	go elevio.PollStopButton(event.stop)

	//turn off all lights
	for f := 0; f < numFloors; f++ {
		for b := elevio.ButtonType(0); b < 3; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}

	//go to known state (down to closest floor)
	elevio.SetMotorDirection(d)

	floor = <-event.floors
	d = elevio.MD_Stop
	elevio.SetMotorDirection(d)
	return event, floor
}


func UpdateLights(orders structs.Order_com) {
	for {
		select {
		case a := <-orders.Light:
			elevio.SetButtonLamp(a.Button, a.Floor, a.Value)
		}
	}
}


func SendButtonPresses(orders structs.Order_com, event Event) {
	for {
		a := <- event.buttons
		orders.OrderFromButton <- structs.Order{Floor: a.Floor, Button: a.Button}
	}
}


func ShouldStop(localOrders [][]int, currentFloor int, lastDir int, maxFloors int) (bool) {
	if (lastDir == 1) {
		if (localOrders[currentFloor][0] == 1 || localOrders[currentFloor][2] == 1) {
			return true
		}
		if (currentFloor == maxFloors - 1 && localOrders[currentFloor][1] == 1) {
			return true
		}
		if (!OrdersInDirection(lastDir, localOrders, currentFloor, maxFloors)) {
			return true
		}
	}
	if (lastDir == -1) {
		if (localOrders[currentFloor][1] == 1|| localOrders[currentFloor][2] == 1) {
			return true
		}
		if (currentFloor == 0 && localOrders[currentFloor][0] == 1) {
			return true
		}
		if (!OrdersInDirection(lastDir, localOrders, currentFloor, maxFloors)) {
			return true
		}
	}
	return false
}


func OrdersInDirection(dir int, localOrders [][]int, currentFloor int, maxFloors int) (bool) {
	if (dir == 1) {
		for i := currentFloor + 1; i < maxFloors; i++ {
			if (localOrders[i][0] == 1 || localOrders[i][1] == 1 || localOrders[i][2] == 1) {
				return true
			}
		}
	}

	if (dir == -1) {
		for i := currentFloor - 1; i >= 0; i-- {
			if (localOrders[i][0] == 1 || localOrders[i][1] == 1 || localOrders[i][2] == 1) {
				return true
			}
		}
	}
	return false
}


func UpdateMovement(lastDir *int, localOrders [][]int, currentFloor int, maxFloors int, idle *bool, Update_out_msg_CH chan<- structs.Message_struct, outgoing_msg structs.Message_struct) {
	//fmt.Printf("Updating movement \n")
	oppositeDir := *lastDir * (-1)
	if (OrdersInDirection(*lastDir, localOrders, currentFloor, maxFloors)) {
		elevio.SetMotorDirection(elevio.MotorDirection(*lastDir))
		outgoing_msg.Dir = structs.MotorDirection(*lastDir)
		outgoing_msg.State = 1
		*idle = false

	} else if (OrdersInDirection(oppositeDir, localOrders, currentFloor, maxFloors)) {
		elevio.SetMotorDirection(elevio.MotorDirection(oppositeDir))
		*lastDir = oppositeDir
		outgoing_msg.Dir = structs.MotorDirection(oppositeDir)
		outgoing_msg.State = 1
		*idle = false

	} else {
		elevio.SetMotorDirection(elevio.MD_Stop)
		outgoing_msg.Dir = 0
		outgoing_msg.State = 0
		*idle = true
	}
	go func() { Update_out_msg_CH <- outgoing_msg }()
	//fmt.Printf("Updated movement \n")
}

//always UpdateMovement after ExecuteStop
func ExecuteStop(localOrders [][]int, orders structs.Order_com, currentFloor int, Update_out_msg_CH chan<- structs.Message_struct, outgoing_msg structs.Message_struct) {
	//fmt.Printf("Executing stop at floor %d\n", currentFloor)
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetDoorOpenLamp(true)
	outgoing_msg.State = 2
	go func() { Update_out_msg_CH <- outgoing_msg }()
	time.Sleep(1 * time.Second)
	elevio.SetDoorOpenLamp(false)
	outgoing_msg.State = 0
	go func(){ Update_out_msg_CH <- outgoing_msg }()
	localOrders[currentFloor][0] = 0
	localOrders[currentFloor][1] = 0
	localOrders[currentFloor][2] = 0
	orders.OrderDone <- structs.Order{Floor: currentFloor}

}


func Watchdog(resetTimer <-chan int, resetElevator chan<- int, doneResetting <-chan int, idle *bool) {

	timer := time.NewTimer(10 * time.Second)

	for {
	  select {
		case <-resetTimer:
			//fmt.Println("Timer reset")
			timer.Reset(10 * time.Second)

		case <-timer.C:
			//fmt.Println("Timer ran out, idle: ", *idle)
			if (!(*idle)) {
				//fmt.Println("Resetting elevator")
				resetElevator<- 1
				<-doneResetting
				//fmt.Println("Done resetting")
			}
      timer.Reset(10 * time.Second)
	  }
	}
}


func OperateElev(orders structs.Order_com, event Event, f int, maxFloors int, Update_out_msg_CH chan<- structs.Message_struct, outgoing_msg structs.Message_struct) {

	go UpdateLights(orders)
	go SendButtonPresses(orders, event)

	//localOrders[floor][button], buttons 0 to 2: up, down, cab
	localOrders := make([][]int, maxFloors)
  for i := 0; i < maxFloors; i++ {
		localOrders[i] = make([]int, 3)
	}

  currentFloor := f
	lastDir := elevio.MD_Down

  //UpdateMovement
  //ExecuteStop := make(chan int, 4096)
	idle := true

  //watchdog com
	resetTimer := make(chan int, 4096)
	resetElevator := make(chan int, 4096)
	doneResetting := make(chan int, 4096)

  go Watchdog(resetTimer, resetElevator, doneResetting, &idle)


  for {
		select {
		case order := <- orders.OrderForLocal:
      resetTimer <- 1
			orders.Light <- structs.LightOrder{Floor: order.Floor, Button: order.Button, Value: true}

			localOrders[order.Floor][order.Button] = 1
			UpdateMovement(&lastDir, localOrders, currentFloor, maxFloors, &idle, Update_out_msg_CH, outgoing_msg)
			//fmt.Println(localOrders)
			outgoing_msg.Queue[order.Floor][order.Button] = 1
			go func() { Update_out_msg_CH <- outgoing_msg }()
			if (idle == true && order.Floor == currentFloor) {
				  ExecuteStop(localOrders, orders, currentFloor, Update_out_msg_CH, outgoing_msg)
			}
			UpdateMovement(&lastDir, localOrders, currentFloor, maxFloors, &idle, Update_out_msg_CH, outgoing_msg)

		case floor := <-event.floors:
			outgoing_msg.Last_floor = floor
			resetTimer <- 1
			go func() { Update_out_msg_CH <- outgoing_msg }()
			currentFloor = floor
			elevio.SetFloorIndicator(floor)
			if ShouldStop(localOrders, currentFloor, lastDir, maxFloors) {
				//ExecuteStop <- 1
				ExecuteStop(localOrders, orders, currentFloor, Update_out_msg_CH, outgoing_msg)

				UpdateMovement(&lastDir, localOrders, currentFloor, maxFloors, &idle, Update_out_msg_CH, outgoing_msg)
			} else {
				//UpdateMovement <- 1
				UpdateMovement(&lastDir, localOrders, currentFloor, maxFloors, &idle, Update_out_msg_CH, outgoing_msg)
			}

			case <-resetElevator:
				outgoing_msg.State = 3
				go func(){ Update_out_msg_CH <- outgoing_msg }()
				recoveryDir := elevio.MotorDirection(1)
				if (currentFloor == 0) {
					recoveryDir = elevio.MotorDirection(1)
				} else {
					recoveryDir = elevio.MotorDirection(-1)
				}
				elevio.SetMotorDirection(recoveryDir)
				recoveryFloor := <-event.floors
				/*
				L:
				for {
					select {
					case <-event.floors:
						break L

					default:
						//keep trying in case motor stops again
						elevio.SetMotorDirection(recoveryDir)
					}
				}
				*/
				elevio.SetMotorDirection(elevio.MD_Stop)
				//fmt.Println("***Managed to stop")
				doneResetting <- 1
				//fmt.Println("***doneResetting sent")

				/*
				event.floors <- recoveryFloor
				fmt.Println("***recoveryFloor sent")
        */

				outgoing_msg.Last_floor = recoveryFloor
				go func() { Update_out_msg_CH <- outgoing_msg }()
				currentFloor = recoveryFloor
				elevio.SetFloorIndicator(recoveryFloor)

				outgoing_msg.State = 0
				go func(){ Update_out_msg_CH <- outgoing_msg }()
				//fmt.Println("Recovery status sent")
				event.floors <- recoveryFloor
		}
	}
}
