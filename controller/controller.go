package controller


import (
	"fmt"
  //"strconv"

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
    floors:     make(chan int),
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
	fmt.Println("Initiating elev\n")
  elevio.SetMotorDirection(d)

  floor = <-event.floors
  d = elevio.MD_Stop
  elevio.SetMotorDirection(d)
  fmt.Printf("Elevator initiated at floor %d \n", floor)
  return event, floor
}


func UpdateLights(orders structs.Order_com) {
  for {
		select {
		case a := <-orders.Light:
      elevio.SetButtonLamp(a.Button, a.Floor, a.Value)
			//fmt.Printf("Lamp on: button %d floor %d \n", a.Button, a.Floor)
		}
	}
}


func SendButtonPresses(orders structs.Order_com, event Event) {
	for {
		a := <- event.buttons
	  orders.OrderFromButton <- structs.Order{Floor: a.Floor, Button: a.Button}
    //fmt.Printf("Button sent %d %d \n", a.Button, a.Floor)
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


func Operate_elev(orders structs.Order_com, event Event, f int, maxFloors int, Update_out_msg_CH chan<- structs.Message_struct, outgoing_msg structs.Message_struct) {

  go UpdateLights(orders)
	go SendButtonPresses(orders, event)
  fmt.Println("Update lights and send button presses \n")

	//localOrders[floor][button]
	//buttons 0 to 2: up,down,cab
	localOrders := make([][]int, maxFloors)
  for i := 0; i < maxFloors; i++ {
		localOrders[i] = make([]int, 3)
	}

  currentFloor := f
	lastDir := elevio.MD_Down

  updateMovement := make(chan int, 4096)
  executeStop := make(chan int, 4096)
	idle := true

  for {
		select {
		case order := <- orders.OrderForLocal:
			fmt.Printf("Updating orders\n")
		  localOrders[order.Floor][order.Button] = 1
			updateMovement <- 1
			fmt.Printf("Updated order\n")
			fmt.Println(localOrders)
			outgoing_msg.Queue = localOrders
			go func() { Update_out_msg_CH <- outgoing_msg }()
			if (idle == true && order.Floor == currentFloor) {
				executeStop <- 1
				//executeStop(localOrders, orders, currentFloor)
			}
			//updateMovement(lastDir, localOrders, currentFloor, maxFloors, &idle)

		case floor := <-event.floors:
        outgoing_msg.Last_floor = floor
        go func() { Update_out_msg_CH <- outgoing_msg }()
        currentFloor = floor
				elevio.SetFloorIndicator(floor)
				fmt.Printf("Reached floor %d", currentFloor)
        if ShouldStop(localOrders, currentFloor, lastDir, maxFloors) {
					executeStop <- 1
					//executeStop(localOrders, orders, currentFloor)
					//updateMovement(lastDir, localOrders, currentFloor, maxFloors, &idle)
				} else {
					updateMovement <- 1
					//updateMovement(lastDir, localOrders, currentFloor, maxFloors, &idle)
				}

		case <-updateMovement:
		  //check for orders in current dir
			//(check for orders in other dir)
			//go towards order
			//or idle
			fmt.Printf("Updating movement \n")
			oppositeDir := lastDir * (-1)
			if (OrdersInDirection(lastDir, localOrders, currentFloor, maxFloors)) {
			  elevio.SetMotorDirection(elevio.MotorDirection(lastDir))
				outgoing_msg.Dir = structs.MotorDirection(lastDir)
				outgoing_msg.State = 1
				idle = false

			} else if (OrdersInDirection(oppositeDir, localOrders, currentFloor, maxFloors)) {
				elevio.SetMotorDirection(elevio.MotorDirection(oppositeDir))
				lastDir = oppositeDir
				outgoing_msg.Dir = structs.MotorDirection(oppositeDir)
        outgoing_msg.State = 1
				idle = false

			} else {
				elevio.SetMotorDirection(elevio.MD_Stop)
				outgoing_msg.Dir = 0
        outgoing_msg.State = 0
				idle = true
			}
			go func() { Update_out_msg_CH <- outgoing_msg }()
			fmt.Printf("Updated movement \n")


		case <-executeStop:
			fmt.Printf("Executing stop at floor %d\n", currentFloor)
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			outgoing_msg.State = 2
			time.Sleep(1 * time.Second)
			elevio.SetDoorOpenLamp(false)
			go func() { Update_out_msg_CH <- outgoing_msg }()
			localOrders[currentFloor][0] = 0
			localOrders[currentFloor][1] = 0
			localOrders[currentFloor][2] = 0
			orders.OrderDone <- structs.Order{Floor: currentFloor}
			updateMovement <- 1
			fmt.Printf("Done executing floor %d\n", currentFloor)

		}
	}
}
