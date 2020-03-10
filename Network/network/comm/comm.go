package comm

import (
	"time"
	"fmt"
	"../peers"
  ."../../../structs"
)


type Communication_ch struct{
  Peer_Update_CH              chan peers.PeerUpdate
  New_peer_CH                 chan string
  Lost_peers_CH               chan []string
  Init_out_msg_CH             chan Message_struct
  Update_out_msg_CH           chan Message_struct
  Out_msg_CH                  chan Message_struct
  Incoming_msg_CH             chan Message_struct
  Update_control_CH           chan Message_struct
}

var out_msg Message_struct

func Comm(comm Communication_ch){
  tick_bcast := time.NewTicker(50*time.Millisecond)
  for{
          select{
          case p:= <-comm.Peer_Update_CH:
            fmt.Printf("Peer update:\n")
      			fmt.Printf("  Peers:    %q\n", p.Peers)
      			fmt.Printf("  New:      %q\n", p.New)
      			fmt.Printf("  Lost:     %q\n", p.Lost)

            if len(p.New) > 0{
                go func(){comm.New_peer_CH <- p.New}()
              }

            if len(p.Lost) > 0{
                go func(){comm.Lost_peers_CH <- p.Lost}()
            }

          case out_msg = <-comm.Init_out_msg_CH:

          case incoming_msg := <-comm.Incoming_msg_CH:
              go func(){comm.Update_control_CH <- incoming_msg}()

          case out_msg = <-comm.Update_out_msg_CH:

          case <-tick_bcast.C:
              go func(){comm.Out_msg_CH <- out_msg}()


          }
    }
}
