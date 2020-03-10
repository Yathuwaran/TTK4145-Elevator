package com_chunication_handler

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

func Communication_handler(com_ch Communication_ch){
  tick_bcast := time.NewTicker(50*time.Millisecond)
  for{
          select{
          case p:= <-com_ch.Peer_Update_CH:
            fmt.Printf("Peer update:\n")
      			fmt.Printf("  Peers:    %q\n", p.Peers)
      			fmt.Printf("  New:      %q\n", p.New)
      			fmt.Printf("  Lost:     %q\n", p.Lost)

            if len(p.New) > 0{
                go func(){com_ch.New_peer_CH <- p.New}()
              }

            if len(p.Lost) > 0{
                go func(){com_ch.Lost_peers_CH <- p.Lost}()
            }

          case out_msg = <-com_ch.Init_out_msg_CH:

          case incoming_msg := <-com_ch.Incoming_msg_CH:
              go func(){com_ch.Update_control_CH <- incoming_msg}()

          case out_msg = <-com_ch.Update_out_msg_CH:

          case <-tick_bcast.C:
              go func(){com_ch.Out_msg_CH <- out_msg}()




          }
    }
}
