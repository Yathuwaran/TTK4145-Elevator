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

func ShouldStop(localOrders [][]int, currentFloor int, lastDir int) (bool) {
	if (lastDir == 1) {
		if (localOrders[currentFloor][0] == 1 || localOrders[currentFloor][2] == 1) {
			return true
		}
	}
	if (lastDir == -1) {
		if (localOrders[currentFloor][1] == 1|| localOrders[currentFloor][2] == 1) {
			return true
		}
	}
	return false
}

func OrdersInDirection(dir int, localOrders [][]int, currentFloor int, maxFloors int) (bool) {
	if (dir == 1) {
	  for i := currentFloor + 1; i < maxFloors; i++ {
		  if (localOrders[i][0] == 1 || localOrders[i][2] == 1) {
				return true
			}
	  }
	}

	if (dir == -1) {
	  for i := currentFloor - 1; i >= 0; i-- {
		  if (localOrders[i][1] == 1 || localOrders[i][2] == 1) {
				return true
			}
	  }
	}
	return false
}

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

  updateMovement := make(chan int, 2048)
  executeStop := make(chan int, 2048)

  for {
		select {
		case order := <- orders.OrderForLocal:
			fmt.Printf("Updating orders %d", order.Button, order.Floor)
		  localOrders[order.Floor][order.Button] = 1
			updateMovement <- 1
			fmt.Printf("Updated order \n")
			fmt.Println(localOrders)

		case floor := <-event.floors:
        currentFloor = floor
				fmt.Printf("Reached floor %d", currentFloor)
        if ShouldStop(localOrders, currentFloor, lastDir) {
					executeStop <- 1
				} else {
					updateMovement <- 1
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

			} else if (OrdersInDirection(oppositeDir, localOrders, currentFloor, maxFloors)) {
				elevio.SetMotorDirection(elevio.MotorDirection(oppositeDir))
				lastDir = oppositeDir
			} else {
				elevio.SetMotorDirection(elevio.MD_Stop)
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


func main(){
  event, current_floor := Init_elev(15657, 4)

  //this should be made in the real main and sent to the order module as well
  orders := structs.Order_com{
    OrderFromButton: make(chan structs.Order),
		OrderForLocal: make(chan structs.Order),
    OrderDone: make(chan structs.Order),
	  Light: make(chan structs.LightOrder)}

  numFloors := 4
  go Fake_gen_orders(orders, event)
	fmt.Println("Initiated fake orders")
  go Operate_elev(orders, event, current_floor, numFloors)
  fmt.Println("Initiated operate elev")
  for{}
}


/*
func Elev_status() {}


func motor_move(dir) {}
func motor_stop() {}
func motor_status() {}

func light_on(light_num) {}
func light_off(light_num) {}
func check_light_status(light_num) {}
*/
