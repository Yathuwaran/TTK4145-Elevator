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

type Order_com struct {
  order      chan structs.Order
  order_done chan structs.Order
}

func Init_elev(port int, numFloors int) (Event, elevio.MotorDirection, int) {
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

  //go to known state (down to closest floor)
  elevio.SetMotorDirection(d)
  floor = <-event.floors
  d = elevio.MD_Stop
  elevio.SetMotorDirection(d)

  return event, d, floor
}

func Fake_gen_orders(orders Order_com, event Event) {
  for {
    select {
    case a := <- event.buttons:
      orders.order <- structs.Order{Floor: a.Floor,
                        Button: a.Button}
    case b := <-orders.order_done:
      fmt.Println("Order complete: ", b.Floor)
    }
  }
}

func Operate_elev(orders Order_com, event Event, d elevio.MotorDirection, f int) {
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
  current_floor := f
  floor_goal := -1
  for {
    select {
    case a := <-orders.order:
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
        orders.order_done <- structs.Order{Floor: b}
      }
    }
  }
}


func main(){
  event, dir, floor := Init_elev(15657, 4)

  //this should be made in the real main and sent to the order module as well
  orders := Order_com{
    order: make(chan structs.Order),
    order_done: make(chan structs.Order)}

  go Fake_gen_orders(orders, event)
  go Operate_elev(orders, event, dir, floor)

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
