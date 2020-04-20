First initiate a Simulator by writing "./SimElevatorServer --port PORTNR"

Then start main by writing "go run main.go PORTNR"

The elevator is programmed such that it works independently from the others. If one were to disconnect, it would behave as though the others didn't exist and vice versa. This is done by continuously sending "I'm alive" messages.


The controller module has no "online" responsibility and only makes sure the local elevator behaves properly, with orders given by the order module. This includes an init to put the elevator in a known state, a watchdog timer to reset the elevator if it behaves incorrectly, and the core for-select loop. This modules also utilizes elevio to set up routines to send order information (button presses).
