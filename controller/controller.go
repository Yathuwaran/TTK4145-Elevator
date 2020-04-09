package main

import (
	"fmt"
  "strconv"

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
/*
ELEVATOR DOESNT OPEN WHEN BUTTONS ON CURRENT FLOOR
*/

func Fake_gen_orders(orders structs.Order_com, event Event) {
  for {
    select {
    case a := <- orders.OrderFromButton:
      orders.OrderForLocal <- structs.Order{Floor: a.Floor,
												Button: a.Button}
			orders.Light <- structs.LightOrder{Floor: a.Floor,
												Button: a.Button, Value: true}
      fmt.Printf("Order module got order at floor %d \n", a.Floor)
    case b := <-orders.OrderDone:
      fmt.Println("Order complete: ", b.Floor)
			orders.Light <- structs.LightOrder{Floor: b.Floor,
												Button: 0, Value: false}
      orders.Light <- structs.LightOrder{Floor: b.Floor,
												Button: 1, Value: false}
			orders.Light <- structs.LightOrder{Floor: b.Floor,
												Button: 2, Value: false}
    }
  }
}



func Init_elev(port int, numFloors int) (Event, int) {
  elevio.Init("localhost:"+strconv.Itoa(port), numFloors)

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
	fmt.Println("Initiating elev")
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
			fmt.Printf("Lamp on: button %d floor %d \n", a.Button, a.Floor)
		}
	}
}

func SendButtonPresses(orders structs.Order_com, event Event) {
	for {
		a := <- event.buttons
	  orders.OrderFromButton <- structs.Order{Floor: a.Floor, Button: a.Button}
    fmt.Printf("Button sent %d %d", a.Button, a.Floor)
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
		  if (localOrders[i][0] == 1 || localOrders[i][2] == 1 || localOrders[i][1] == 1) {
				return true
			}
	  }
	}

	if (dir == -1) {
	  for i := currentFloor - 1; i >= 0; i-- {
		  if (localOrders[i][1] == 1 || localOrders[i][2] == 1 || localOrders[i][0] == 1) {
				return true
			}
	  }
	}
	return false
}
/*
func updateMovement(lastDir int, localOrders [][]int, currentFloor int, maxFloors int, idle *bool) {
	oppositeDir := lastDir * (-1)
	if (OrdersInDirection(lastDir, localOrders, currentFloor, maxFloors)) {
		elevio.SetMotorDirection(elevio.MotorDirection(lastDir))
		*idle = false

	} else if (OrdersInDirection(oppositeDir, localOrders, currentFloor, maxFloors)) {
		elevio.SetMotorDirection(elevio.MotorDirection(oppositeDir))
		lastDir = oppositeDir
		*idle = false

	} else {
		elevio.SetMotorDirection(elevio.MD_Stop)
		*idle = true
	}
}


func executeStop(localOrders [][]int, orders structs.Order_com, currentFloor int) {
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetDoorOpenLamp(true)
	time.Sleep(1 * time.Second)
	elevio.SetDoorOpenLamp(false)
	localOrders[currentFloor][0] = 0
	localOrders[currentFloor][1] = 0
	localOrders[currentFloor][2] = 0
	orders.OrderDone <- structs.Order{Floor: currentFloor}
}
*/

func Operate_elev(orders structs.Order_com, event Event, f int, maxFloors int) {

  go UpdateLights(orders)
	go SendButtonPresses(orders, event)
  fmt.Println("updatelights and sendbuttonpresses")

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
			fmt.Printf("Updating orders %d", order.Button, order.Floor)
		  localOrders[order.Floor][order.Button] = 1
			updateMovement <- 1
			fmt.Printf("Updated order \n")
			fmt.Println(localOrders)
			if (idle == true && order.Floor == currentFloor) {
				executeStop <- 1
				//executeStop(localOrders, orders, currentFloor)
			}
			//updateMovement(lastDir, localOrders, currentFloor, maxFloors, &idle)

		case floor := <-event.floors:
        currentFloor = floor
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
				idle = false

			} else if (OrdersInDirection(oppositeDir, localOrders, currentFloor, maxFloors)) {
				elevio.SetMotorDirection(elevio.MotorDirection(oppositeDir))
				lastDir = oppositeDir
				idle = false

			} else {
				elevio.SetMotorDirection(elevio.MD_Stop)
				idle = true
			}
			fmt.Printf("Updated movement \n")


		case <-executeStop:
			fmt.Printf("Executing stop at floor %d", currentFloor)
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			time.Sleep(1 * time.Second)
			elevio.SetDoorOpenLamp(false)
			localOrders[currentFloor][0] = 0
			localOrders[currentFloor][1] = 0
			localOrders[currentFloor][2] = 0
			orders.OrderDone <- structs.Order{Floor: currentFloor}
			updateMovement <- 1
			fmt.Printf("Done executing floor %d", currentFloor)

		}
	}
}
