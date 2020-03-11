package main

import(
  "./network/bcast"
  "./network/localip"
	"./network/peers"
  com"./network/communication_handler"
  ."../structs"
  "fmt"
  "os"
  "time"

)


func main() {

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
    Update_out_msg_CH:            make(chan Message_struct),
    Out_msg_CH:                   make(chan Message_struct),
    Incoming_msg_CH:              make(chan Message_struct),
    Update_control_CH:            make(chan Message_struct)}

    opp := os.Args[1]
    //Dummy struct
    var mld Message_struct
    mld.ID = opp
    mld.Destination = Order{1,1}
    mld.Last_floor = 2
    mld.Dir = 1
    mld. State = 2
    mld.Queue = [][]int{{0, 1},{6, 7}}

    peer_ch:= make(chan bool)

    go func(){
      for{
        test.Update_out_msg_CH <-mld
        mld.Last_floor++
        time.Sleep(50 * time.Millisecond)
      }
    }()
    go func(){for
      {
        a:= <-test.Update_control_CH
          fmt.Println(a)
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
