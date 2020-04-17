package orderModule

import (
	S"../../structs"
	"math"
	"fmt"
	D"../../elevio"

	"sort"
	"sync"
	"time"
)

var mutex = &sync.Mutex{}

func GenerateOrders(orders S.Order_com, ExternalOrder chan S.ExternalOrder, elevArray map[string]*S.Message_struct, id string, outgoing_msg S.Message_struct,Update_out_msg_CH chan<- S.Message_struct) {
  for {
    select {
    case a := <- orders.OrderFromButton:
			if (int(a.Button) == 2 ){
				orders.OrderForLocal <- a
			}else{
			external, status := GetBestElev(elevArray, a)
			if(status){
			fmt.Println("The best elevator is:", external.ID, "\n")
			ExternalOrder <- external
		}else{fmt.Println("Order Already taken")}
		}
    case b := <-orders.OrderDone:
      fmt.Println("Order complete \n")
			for i := 0; i < 3; i++ {
				if(outgoing_msg.Queue[b.Floor][i] == 1){

					orders.Light <- S.LightOrder{Floor: b.Floor,
												Button: D.ButtonType(i), Value: false}
					outgoing_msg.Queue[b.Floor][i] = 0
			}
		}
		go func(){ Update_out_msg_CH <- outgoing_msg }()

     case c:= <- ExternalOrder:
			 if (c.ID) == id{
				 fmt.Println("---------------------I'll handle the order-----------------------")
				 orders.OrderForLocal <- c.Destination
				 outgoing_msg.ExternalOrder.ID = "Empty"
    		} else {
					outgoing_msg.ExternalOrder = c
					fmt.Println("-----------------------Sending order-----------------------")
					go func() { Update_out_msg_CH <- outgoing_msg
											time.Sleep(100 * time.Millisecond)
											outgoing_msg.ExternalOrder.ID = "Empty"
											Update_out_msg_CH <- outgoing_msg
											}()
				}


  }
}
}


func LightOrders (orders S.Order_com, elevArray map[string]*S.Message_struct, numFloors int, id string, key []string, externalLight [][]int) ([][]int) {
		//fmt.Println(key, "LIGHTS")
		for i := 0; i < len(key); i++ {
			Queue:=elevArray[key[i]].Queue
			if (((elevArray[key[i]]).ID != id) && len(Queue) >0){
				for  j := 0; j < numFloors; j++{
					for k := 0; k < 2; k++{
						if((Queue[j][k]== 1) && (externalLight[j][k] == 0) ){
						orders.Light <- S.LightOrder{Floor: j, Button: D.ButtonType(k), Value: true}
						externalLight[j][k] = 1
						}else if((Queue[j][k]== 0) && (externalLight[j][k] == 1) ){
						orders.Light <- S.LightOrder{Floor: j, Button: D.ButtonType(k), Value: false}
						externalLight[j][k] = 0

						}
				}
			}
		}
	}
	return externalLight
}

func AddExternalOrder(orders S.Order_com, elevArray map[string]*S.Message_struct, id string,key[]string){

		for i := 0; i < len(key); i++ {
				if ((elevArray[key[i]].ExternalOrder).ID == id){

					fmt.Println("---------------Received External Order---------------")
					orders.OrderForLocal <- elevArray[key[i]].ExternalOrder.Destination
					mutex.Lock()
					((*elevArray[key[i]]).ExternalOrder).ID = "Empty"
					mutex.Unlock()
			}
		}

}


func UpdatePeers(elevArray map[string]*S.Message_struct, New_peer_CH <- chan string, Lost_peers_CH <- chan []string, orders S.Order_com, Update_control_CH <- chan  S.Message_struct){
	for{
		select{
		case a:= <- New_peer_CH:
			var init_msg S.Message_struct
			mutex.Lock()
			elevArray[a] = &init_msg
			elevArray[a].ID = a
			mutex.Unlock()

			fmt.Println("New peer added to elevArray: ",elevArray[a],"\n")

		case lost:= <- Lost_peers_CH:
			fmt.Println("-------------LOST QUEUE RECEIVED--------------")
			fmt.Println(elevArray[string(lost[0])].Queue)
			for i := 0; i < len(elevArray[string(lost[0])].Queue); i++{
				for j := 0; j < len(elevArray[string(lost[0])].Queue[i]); j++{
					if (elevArray[string(lost[0])]).Queue[i][j] == 1{
						orders.OrderFromButton <- S.Order{Floor: i, Button: D.ButtonType(j)}
					}
				}
			}
			mutex.Lock()
			delete(elevArray,lost[0])
			mutex.Unlock()
				}
		}
}

func UpdateElevArray(elevArray map[string]*S.Message_struct, update  S.Message_struct) {
	//Update elevator_list from msg
		if elevArray[update.ID] == nil {
			return
		} else{

		mutex.Lock()
		(*elevArray[update.ID]).ID = update.ID
		(*elevArray[update.ID]).Last_floor = update.Last_floor
		(*elevArray[update.ID]).Dir = update.Dir
		(*elevArray[update.ID]).State = update.State
		(*elevArray[update.ID]).Queue = update.Queue
		(*elevArray[update.ID]).ExternalOrder = update.ExternalOrder
		mutex.Unlock()
	}
}

// GetBestElev returns the best suited elevator for an order
func GetBestElev(elevArray map[string]*S.Message_struct, order S.Order)  (S.ExternalOrder, bool) {
	bestIndex := 0
	bestScore := 0
	counter   := 0
	status		:= false

	mutex.Lock()
	var keys []string
	for k:= range elevArray{
		keys = append(keys,k)
	}
	mutex.Unlock()
	sort.Strings(keys)

	for i := 0; i < len(elevArray) ; i++ {
		score := ElevatorCostFunction(*elevArray[keys[i]], order)
		fmt.Printf("Elevator: %s     Score: %d\n", elevArray[keys[i]].ID, score)
		if score > bestScore {
			bestScore = score
			bestIndex = i
		}
if (((elevArray[keys[i]]).Queue[order.Floor][order.Button]) == 0){
			counter ++
		}
	}
	if (counter == len(elevArray)){status = true}else{status = false}
	return S.ExternalOrder{ID: keys[bestIndex], Destination: order}, status
}

// ElevatorCostFunction determines how well suited an elevator is for a particular order
func ElevatorCostFunction(peer S.Message_struct, order S.Order) int {
	var totalScore int
	const ( // These values might need tweaking
		Weight_PickupOnPassing       = 50
		Weight_DistanceToOrder       = 10
		Weight_ActivityStateRunning  = 1
		Weight_ActivityStateIdle     = 50
		Weight_ActivityStateOpenDoor = 45
	)

	// A high weight-score means that the evaluated elevator is well suited for the order

	///////////////////////////////////////////////////////////////////////////////

	// If going down, and new order is on the way to existing order
	if (peer.Dir == S.MD_Down) && (order.Floor < peer.Last_floor) && ((order.Button == S.BT_Cab) || (order.Button == S.BT_HallDown)) {

		totalScore += (1 * Weight_PickupOnPassing) // This should weigh a lot

	}

	///////////////////////////////////////////////////////////////////////////////

	// If going up, and new order is on the way to existing order
	if (peer.Dir == S.MD_Up) && (order.Floor > peer.Last_floor) && ((order.Button == D.BT_Cab) || (order.Button == D.BT_HallUp)) {

		totalScore += (1 * Weight_PickupOnPassing) // This should weigh a lot

	}

	///////////////////////////////////////////////////////////////////////////////

	floorDifference := int(math.Abs(float64(order.Floor - peer.Last_floor)))
	totalScore += (4 - floorDifference) * Weight_DistanceToOrder

	///////////////////////////////////////////////////////////////////////////////

	if peer.State == S.Moving {

		totalScore += (1 * Weight_ActivityStateRunning)

	} else if peer.State == S.IDLE {

		totalScore += (1 * Weight_ActivityStateIdle)

	} else if peer.State == S.LOST_CONN {
		totalScore -= 10000
	} else if peer.State == S.OPENDOOR {
		totalScore += (1 * Weight_ActivityStateOpenDoor)
	}

	///////////////////////////////////////////////////////////////////////////////

	return totalScore

}
