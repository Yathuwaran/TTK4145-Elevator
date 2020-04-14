package main

import(
  "fmt"
  ."./structs"
  "./network/network/bcast"
  "./network/network/localip"
	"./network/network/peers"
  com"./network/network/communication_handler"
  "./order/orderModule"
  "./controller"
  "os"


)

func main(){

port := os.Args[1]
numFloors := 4

type elevArr map[string]*Message_struct

ExternalOrders := make(chan ExternalOrder, 4096)
elevArray := make(elevArr)
peer_ch:= make(chan bool)
out := make(chan Message_struct)
trigger := make(chan int)

var outgoing_msg Message_struct
var id string
var peer_list []string

Queue := make([][]int, numFloors)
for i := 0; i < numFloors; i++ {
  Queue[i] = make([]int, 3)
  for j := 0; j < 3; j++{
    Queue[i][j] = 0
  }
}

externalLight := make([][]int, numFloors)
for i := 0; i < numFloors; i++ {
  externalLight[i] = make([]int, 3)
  for j := 0; j < 3; j++{
    externalLight[i][j] = 0
  }
}


if id == "" {
    localIP, err := localip.LocalIP()
    if err != nil {
      fmt.Println(err)
      localIP = "DISCONNECTED"
    }
    id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
  }

outgoing_msg.ID = id
outgoing_msg.Queue = Queue



//Network Init
  comm := com.Communication_ch{
    Peer_Update_CH:               make(chan peers.PeerUpdate),
    New_peer_CH:                  make(chan string),
    Lost_peers_CH:                make(chan []string),
    Update_out_msg_CH:            make(chan Message_struct),
    Out_msg_CH:                   make(chan Message_struct),
    Incoming_msg_CH:              make(chan Message_struct),
    Update_control_CH:            make(chan Message_struct),
    Peers:                        make(chan []string)}


//Controller Init
  event, current_floor := controller.Init_elev(port, 4)

  orders := Order_com{
    OrderFromButton: make(chan Order, 4096),
    OrderForLocal: make(chan Order, 4096),
    OrderDone: make(chan Order, 4096),
    Light: make(chan LightOrder, 4096)}


  go func(){comm.Update_out_msg_CH <- outgoing_msg}()
  go func(){for{peer_list=<-comm.Peers}}()
  go orderModule.UpdatePeers(elevArray, comm.New_peer_CH, comm.Lost_peers_CH, orders, comm.Update_control_CH)
  go orderModule.GenerateOrders(orders, ExternalOrders, elevArray,id, outgoing_msg, out)
  go controller.Operate_elev(orders, event, current_floor, numFloors, out, outgoing_msg)


  go func(){
    for{
      select{
      case a:= <- comm.Update_control_CH:
      orderModule.UpdateElevArray(elevArray,a)
      trigger <- 1
    case b:= <- out:
      go func() { comm.Update_out_msg_CH <- b }()
        }
      }
    }()

  go func () {
    for{
      select{
      case <-trigger:
        go orderModule.AddExternalOrder(orders, elevArray, id, peer_list)
        go func (){externalLight = orderModule.LightOrders(orders, elevArray, numFloors, id, peer_list, externalLight)}()
      }
    }
  }()

//Communication
  go com.Communication_handler(comm)
  go bcast.Transmitter(12345, comm.Out_msg_CH)
  go bcast.Receiver(12345, comm.Incoming_msg_CH)
  go peers.Transmitter(15647, id, peer_ch)
  go peers.Receiver(15647, comm.Peer_Update_CH)

  select{}
}
