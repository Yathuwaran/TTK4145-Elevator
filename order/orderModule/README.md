## Order Module usage

**EnqueueLocal** and **RetrieveLocalOrder** enqueues and dequeues pending orders on local system.

**ElevatorCostFunction** puts a score to an Elevator/Order pair. A high score means the elevator is well suited to handle the order, while a low score means the opposite.

**GetBestElev** is used to determine the best suited elevator from a set of elevators. It takes an array of elevators as input, as well as an order. It then returns the elevator with the highest ElevatorCostFunction-score.
