

This is the project repository for the elevator project in [TTK4145](https://github.com/TTK4145) by [Yathuwaran Raveendranathan](https://github.com/yathuwaran), [Mehmed Adzemovic](https://github.com/mehmeda) and [Hans Olav Lofstad](https://github.com/SupremeAckbar)

First initiate a Simulator by writing "./SimElevatorServer --port PORTNR"

Then start main by writing "go run main.go PORTNR"



TODO:

 - [ ] Fix bugs in controller
 - [ ] Add functionality such that Hall_btn lights are same for all elevs
 - [ ] Add backup (not sure if needed?)

Known bugs

 - [x] Cab orders are sometimes executed by other elevators - FIXED
 - [ ] Some orders are weighed poorly, orders inefficient
