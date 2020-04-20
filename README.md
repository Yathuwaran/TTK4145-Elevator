First initiate a Simulator by writing "./SimElevatorServer --port PORTNR"
Then start main by writing "go run main.go PORTNR"


### Controller
The controller module has no "online" responsibility and only makes sure the local elevator behaves properly, with orders given by the order module. This includes an init to put the elevator in a known state, a watchdog timer to reset the elevator if it behaves incorrectly, and the core for-select loop. This modules also utilizes elevio to set up routines to send order information (button presses).

### Orders and network
Each order is registered locally and then sent through a cost-function to find out which elevator is best suited for the task. There is no master/slave system; Each elevator runs it's own cost-function and distribution of orders, whether it be locally or across the network. If an elevator is offline it is treated as if it doesn't exist to the others, and vice versa. This means that an offline elevator will still operate locally. This is done by continuously sending messages to signal that the sender is alive and functioning. These messages also include the current state of the sender's elevator. To prevent packet loss, the messages are sent multiple times for redundancy.

The following modules were not authored by us, but ```network``` was altered with the ```communication_handler``` files to better suit our application:

* elevio
* network 
