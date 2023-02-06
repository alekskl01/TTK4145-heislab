package elevatorfsm

import (
	"Elevator/elevio"
	"fmt"
)

func elevator_setReady() //	INITIAL ELEVATOR STATE
func getOrder()          //	CONTINUALLY CHECKS BUTTONS, STORES ORDERS
func priorityQueue()     //	PRIORITY ALGORITHM FOR ORDER QUEUE
func elevatorActive()    //	EXECUTE ORDERS, IDLE READY STATE IF NOT

type State int

const (
	IDLE       State = 0
	IDLE_READY       = 1
	MOVING           = 2
	DOOR_OPEN        = 3
)

type elevator struct {
	ELEVATOR_STATE State
	int 			FLOOR          
	DIRECTION      Motor_Direction // vet ikke om denne kan funke :P
}

func SetReady(system elevator) {
	fmt.Println("PREPARING ELEVATOR...\n")
	elevio.DefaultInit()	
	for {

	}
}
