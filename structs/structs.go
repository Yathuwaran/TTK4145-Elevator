package structs

import "../elevio"

type ButtonType int
const(BT_HallUp ButtonType  = 0
      BT_HallDown           = 1
      BT_Cab                = 2
      BT_No_Butt            = -1
      )

type MotorDirection int
const(
  MD_Up MotorDirection = 1
  MD_Down         = -1
  MD_Stop         = 0
)

type Order struct {
  Floor         int
  Button        elevio.ButtonType
}

type LightOrder struct {
  Floor         int
  Button        elevio.ButtonType
  Value         bool
}

type Order_com struct {
  orderFromButton chan Order
  orderForLocal   chan Order
  orderDone       chan Order
	light           chan LightOrder
}

type ElevatorStatus struct{
  Destination   Order
  Last_floor    int
  Dir           MotorDirection
  State         ElevatorState
}

type ElevatorState int
const(
  IDLE ElevatorState  = 0
  Moving              = 1
  OPENDOOR            = 2
  LOST_CONN           = 3)

type Message_struct struct{
  ID            string
  Destination   Order
  Last_floor    int
  Dir           MotorDirection
  State         ElevatorState
  Queue         [][]int
}
