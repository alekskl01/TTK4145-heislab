package main

import (
	"Elevator/buttons"
	"Elevator/elevator"
	"Elevator/elevio"
	"Elevator/fsm"
)

func main() {

	elevio.DefaultInit()

	system := elevator.Elevator{}
	elevator.InitializeElevator(&system)
	go buttons.HandleButtonInputs(&system)
	fsm.RunStateMachine(&system)
}
