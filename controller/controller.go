package main

import (
	"fmt"
  "strconv"

	"../elevio"
  "../structs"
  "time"
)

type Drv struct {
  buttons chan elevio.ButtonEvent
  floors chan int
  obstr chan bool
  stop chan bool
}

type Order_com struct {
  order      chan structs.Order
  order_done chan structs.Order
}

func Init_elev(port int, numFloors int) (Drv, elevio.MotorDirection, int) {
  elevio.Init("localhost:"+strconv.Itoa(port), numFloors)

  var d elevio.MotorDirection = elevio.MD_Down
  floor := -1

  drv := Drv{
    buttons:    make(chan elevio.ButtonEvent),
    floors:     make(chan int),
    obstr:      make(chan bool),
    stop:       make(chan bool)}

  go elevio.PollButtons(drv.buttons)
  go elevio.PollFloorSensor(drv.floors)
  go elevio.PollObstructionSwitch(drv.obstr)
  go elevio.PollStopButton(drv.stop)

  //go to known state (down to closest floor)
  elevio.SetMotorDirection(d)
  floor = <-drv.floors
  d = elevio.MD_Stop
  elevio.SetMotorDirection(d)

  return drv, d, floor
}

func Fake_gen_orders(orders Order_com, drv Drv) {
  for {
    select {
    case a := <- drv.buttons:
      orders.order <- structs.Order{Floor: a.Floor,
                        Button: a.Button}
    case b := <-orders.order_done:
      fmt.Println("Order complete: ", b.Floor)
    }
  }
}

func Operate_elev(orders Order_com, drv Drv, d elevio.MotorDirection, f int) {
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

    case b := <-drv.floors:
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
  drv, d, f := Init_elev(15657, 4)

  o := Order_com{
    order: make(chan structs.Order),
    order_done: make(chan structs.Order)}

  go Fake_gen_orders(o, drv)
  go Operate_elev(o, drv, d, f)

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
