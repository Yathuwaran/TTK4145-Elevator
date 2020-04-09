package main

import(
  "fmt"
  ."./structs"
  //"./backup"
  "./network/network/bcast"
  "./network/network/localip"
	"./network/network/peers"
  "./controller"
  O"./order/orderModule"
  com"./network/network/communication_handler"
  "os"

  //"time"
)

func main(){
  port := os.Args[1]
  //Dummy struct


  //Backup TEST
  //-----------------------
  //var elev Message_struct
  //elev = backup.Read_backup()
  //fmt.Println(elev)


//-------------------------


//Network TEST
//-------------------------
var id string
if id == "" {
  localIP, err := localip.LocalIP()
  if err != nil {
    fmt.Println(err)
    localIP = "DISCONNECTED"
  }
  id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
}

fmt.Println(id)

/*var mld Message_struct
mld.ID = id
mld.Destination = Order{1,1}
mld.Last_floor = 2
mld.Dir = 1
mld. State = 2
mld.Queue = [][]int{{0, 1, 0},{1, 1, 1}}
backup.Write_backup(mld)
*/
//Network communication channels
test := com.Communication_ch{
  Peer_Update_CH:               make(chan peers.PeerUpdate),
  New_peer_CH:                  make(chan string),
  Lost_peers_CH:                make(chan []string),
  Update_out_msg_CH:            make(chan Message_struct),
  Out_msg_CH:                   make(chan Message_struct),
  Incoming_msg_CH:              make(chan Message_struct),
  Update_control_CH:            make(chan Message_struct)}

  peer_ch:= make(chan bool)
  var outgoing_msg Message_struct
  event, current_floor := controller.Init_elev(port, 4)
/*  go func(){
    for{
      test.Update_out_msg_CH <-mld
      mld.Last_floor++
      time.Sleep(50 * time.Millisecond)
    }
  }()
*/
  /*go func(){for
    {
      a:= <-test.Update_control_CH
        fmt.Println(a)
      }
  }()
*/

 //elevs :=  make(map[string]Message_struct)

 orders := Order_com{
   OrderFromButton: make(chan Order, 4096),
 	OrderForLocal: make(chan Order, 4096),
   OrderDone: make(chan Order, 4096),
   Light: make(chan LightOrder, 4096)}

 type elevsd map[string]*Message_struct

 elevs := make(elevsd)

numFloors := 4

go controller.UpdatePeers(elevs, test.New_peer_CH, test.Lost_peers_CH, orders)

go func ()  { for {
		select {
    case inc_msg := <-test.Update_control_CH:
    //fmt.Println(inc_msg)
    controller.UpdateElevArray(elevs, inc_msg)

    case a:= <- test.Lost_peers_CH:
      fmt.Println(a[0])
    case r:= <- orders.OrderFromButton:
      fmt.Println("Order from button received; ",r,"\n")
}
}
}()


//go controller.UpdateElevArray(elevs, test.Update_control_CH)



go com.Communication_handler(test)
//Transmitters/Receivers
go bcast.Transmitter(12345, test.Out_msg_CH)
go bcast.Receiver(12345, test.Incoming_msg_CH)
go peers.Transmitter(15647, id, peer_ch)
go peers.Receiver(15647, test.Peer_Update_CH)


outgoing_msg.ID = id


//this should be made in the real main and sent to the order module as well


ExternalOrders := make(chan ExternalOrder)


go O.GenerateOrders(orders, ExternalOrders, elevs,id, outgoing_msg, test.Update_out_msg_CH)
go O.AddExternalOrder(orders, elevs,id)
//fmt.Println("Initiated fake orders")
go controller.Operate_elev(orders, event, current_floor, numFloors, test.Update_out_msg_CH, outgoing_msg )
fmt.Println("Initiated operate elev \n")


select {}

}
