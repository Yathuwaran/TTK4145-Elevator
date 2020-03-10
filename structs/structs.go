package structs

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
  Button        ButtonType
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
