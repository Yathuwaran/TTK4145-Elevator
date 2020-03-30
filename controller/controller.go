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
    case a := <- event.buttons:
      orders.orderForLocal <- structs.Order{Floor: a.Floor,
                        Button: a.Button}
    case b := <-orders.orderDone:
      fmt.Println("Order complete: ", b.Floor)
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
  elevio.SetMotorDirection(d)
  floor = <-event.floors
  d = elevio.MD_Stop
  elevio.SetMotorDirection(d)

  return event, floor
}

func updateLights(orders structs.Order_com) {
  for {
		select {
		case a := <-orders.light:
      SetButtonLamp(a.Button, a.Floor, a.Value)
		}
	}
}

func sendButtonPresses(orders structs.Order_com, event Event) {
	for {
		select {
		case a := <- event.buttons:
			orders.orderFromButton <- structs.Order{Floor: a.Floor,
												Button: a.Button}
	  }
  }
}

func shouldStop(localOrders [][]int, currentFloor int, lastDir MotorDirection) (bool) {
	if (lastDir == elevio.MD_Up) {
		if (localOrders[currentFloor][0] == 1 || localOrders[currentFloor][2] == 1) {
			return true
		}
	}
	if (lastDir == elevio.MD_Down) {
		if (localOrders[currentFloor][1] == 1|| localOrders[currentFloor][2] == 1) {
			return true
		}
	}
	return false
}


func Operate_elev(orders structs.Order_com, event Event, f int, maxFloors int) {
  /*pseudocode:
  create 2d-array for orders.
  if orders in current dir: go to the one furthest away
  elseif orders in other dir: choose furthest
  else stop

  if driving by floor
  check for orders: open/close door, confirm order

  elevator should somehow give its status
  can be done when passing floors
  motor change should also be updated
  both sent on status channel, not sure if this needs its own module
  (harder to implement separate module, but doing it every floor/motor change gives
  irregular updates. Possibly need timestamp, given by status sender!)

  one routine should send button presses to order module
  should also recieve orders and update the local array (with mutex)
  this allows the elevator to be "stupid", only needs to check stuff when it hits a floor

  */



  go updateLights(orders)
	go sendButtonPresses(orders, event)


	//localOrders[floor][button]
	//buttons 0 to 2: up,down,cab
	localOrders := make([][]int, maxFloors)
  for i := 0; i < maxfloors; i++ {
		localOrders[i] = make([]int, 3)
	}

  currentFloor := f
	lastDir := elevio.MD_Down

  updateMovement := make(chan int)
  executeStop := make(chan int)

  for {
		select {
		case order := <- orders.orderForLocal:
		  localOrders[order.Floor][order.Button] = 1
			updateMovement <- 1

		case floor := <-event.floors:
        currentFloor = floor
        if shouldStop(localOrders, currentFloor, lastDir) {
					executeStop <- 1
				}

		case <-updateMovement:
		  //check for orders in current dir
			//(check for orders in other dir)
			//go towards order
			//or idle

		case <-executeStop:




		default:
      //send status update on timer interval
		}
	}


  /*pseudocode:
	for
	select

	floor <-
    -should stop?
		  do it
			send order done
       SetFloorIndicator(int)
			 SetDoorOpenLamp(bool)


	order_from_orders <-
	  update local orders

  */


  current_floor := f
  floor_goal := -1
  for {
    select {
    case a := <-orders.orderFromButton:
      if current_floor < a.Floor {
        elevio.SetMotorDirection(elevio.MD_Up)
      } else if current_floor > a.Floor {
        elevio.SetMotorDirection(elevio.MD_Down)
      }
      floor_goal = a.Floor

    case b := <-event.floors:
      current_floor = b
      if current_floor == floor_goal {
        elevio.SetMotorDirection(elevio.MD_Stop)
        elevio.SetDoorOpenLamp(true)
        time.Sleep(1 * time.Second)
        elevio.SetDoorOpenLamp(false)
        orders.orderDone <- structs.Order{Floor: b}
      }
    }
  }
}


func main(){
  event, current_floor := Init_elev(15657, 4)

  //this should be made in the real main and sent to the order module as well
  orders := structs.Order_com{
    orderFromButton: make(chan structs.Order),
		orderForLocal: make(chan structs.Order),
    orderDone: make(chan structs.Order),
	  light: make(chan structs.LightOrder)}

  numFloors = 4
  go Fake_gen_orders(orders, event)
  go Operate_elev(orders, event, current_floor, numFloors)

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
