package main

import(
  "fmt"
  ."./structs"
  "./backup"
  "./network/network/bcast"
  "./network/network/localip"
	"./network/network/peers"
  com"./network/network/communication_handler"
  "os"
  "time"
)

func main(){

  //Dummy struct
  var mld Message_struct
  mld.ID = "Elevator_1"
  mld.Destination = Order{1,1}
  mld.Last_floor = 2
  mld.Dir = 1
  mld. State = 2
  mld.Queue = [][]int{{0, 1},{6, 7}}

  //Backup TEST
  //-----------------------
  var elev Message_struct
  elev = backup.Read_backup()
  fmt.Println(elev)
  backup.Write_backup(mld)

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

//Network communication channels
test := com.Communication_ch{
  Peer_Update_CH:               make(chan peers.PeerUpdate),
  New_peer_CH:                  make(chan string),
  Lost_peers_CH:                make(chan []string),
  Init_out_msg_CH:              make(chan Message_struct),
  Update_out_msg_CH:            make(chan Message_struct),
  Out_msg_CH:                   make(chan Message_struct),
  Incoming_msg_CH:              make(chan Message_struct),
  Update_control_CH:            make(chan Message_struct)}

  peer_ch:= make(chan bool)

  go func(){
    for {
      select{

      case a:= <-test.Update_control_CH:
        fmt.Println(a)
      mld.Last_floor++

      //The outgoing message should always be passed to "Update_out_msg_CH". Passing it to "Out_msg_CH"
      //will lead to error.
      test.Update_out_msg_CH <-mld
      time.Sleep(50 * time.Millisecond)
    }
    }
  }()


go com.Communication_handler(test)

//Transmitters/Receivers
go bcast.Transmitter(12345, test.Out_msg_CH)
go bcast.Receiver(12345, test.Incoming_msg_CH)
go peers.Transmitter(15647, id, peer_ch)
go peers.Receiver(15647, test.Peer_Update_CH)


select {}


}