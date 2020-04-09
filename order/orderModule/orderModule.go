package orderModule

import (
	S "../../structs"
	"math"
	"fmt"
	D"../../elevio"

	"sort"
	"sync"
	"time"


)


var mutex = &sync.Mutex{}

//OVERFLØDIG
/*
var orderQueue chan S.Order
var initStatus bool

// QueueInit initializes the orderQueue
func QueueInit() {
	if !initStatus {
		initStatus = true
		orderQueue = make(chan S.Order, 30)
	}
}

// EnqueueLocal puts a new order into the orderQueue
func EnqueueLocal(order S.Order) {
	QueueInit()
	orderQueue <- order
}

// RetrieveLocalOrder returns first order in queue
func RetrieveLocalOrder() S.Order {
	order := <-orderQueue
	return order
}
*/


func GenerateOrders(orders S.Order_com, ExternalOrder chan S.ExternalOrder, elevArray map[string]*S.Message_struct, id string, outgoing_msg S.Message_struct,Update_out_msg_CH chan<- S.Message_struct) {
  for {
    select {
    case a := <- orders.OrderFromButton:
			external := GetBestElev(elevArray, a)


			fmt.Println("The best elevator is:", external.ID, "\n")
			go func (){ExternalOrder <- external}()

    case b := <-orders.OrderDone:
      fmt.Println("Order complete: \n", b.Floor)
			orders.Light <- S.LightOrder{Floor: b.Floor,
												Button: 0, Value: false}
      orders.Light <- S.LightOrder{Floor: b.Floor,
												Button: 1, Value: false}
			orders.Light <- S.LightOrder{Floor: b.Floor,
												Button: 2, Value: false}

     case c:= <- ExternalOrder:
			 if (c.ID) == id{
				 fmt.Println("---------------------I'll handle the order-----------------------")
				 orders.OrderForLocal <- c.Destination
				 orders.Light <- S.LightOrder{Floor: c.Destination.Floor,
	 												Button: c.Destination.Button, Value: true}
				 outgoing_msg.ExternalOrder.ID = "Empty"
    		} else {
					outgoing_msg.ExternalOrder = c
					fmt.Println("-----------------------Sending order-----------------------")
					go func() { Update_out_msg_CH <- outgoing_msg
											time.Sleep(100 * time.Millisecond)
											outgoing_msg.ExternalOrder.ID = "Empty"
											Update_out_msg_CH <- outgoing_msg}()
				}


  }
}
}

func AddExternalOrder(orders S.Order_com, elevArray map[string]*S.Message_struct, id string){
	for{
			mutex.Lock()
			var key []string
			for l:= range elevArray{
					key = append(key,l)
					}
			mutex.Unlock()
			sort.Strings(key)

	for i := 0; i < len(key); i++ {
			if ((elevArray[key[i]].ExternalOrder).ID == id){

				fmt.Println("---------------Received External Order---------------")
				orders.OrderForLocal <- elevArray[key[i]].ExternalOrder.Destination
				orders.Light <- S.LightOrder{Floor: elevArray[key[i]].ExternalOrder.Destination.Floor,
												 Button: elevArray[key[i]].ExternalOrder.Destination.Button, Value: true}
				mutex.Lock()
				((*elevArray[key[i]]).ExternalOrder).ID = "Empty"
				mutex.Unlock()


			}

	}
}}


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

		//fmt.Println(elevArray[Update_control_CH.ID])
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
func GetBestElev(elevArray map[string]*S.Message_struct, order S.Order) S.ExternalOrder {
	bestIndex := 0
	bestScore := 0

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
	}
	return S.ExternalOrder{ID: keys[bestIndex], Destination: order}
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
