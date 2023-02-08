package buttons

import (
	"Elevator/elevator"
	"Elevator/elevio"
)

func HandleButtonInputs(elev *elevator.Elevator) {
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	for {
		select {
		case a := <-drv_buttons:
			if !elev.Requests[a.Floor][a.Button] {
				elev.Requests[a.Floor][a.Button] = true
				elevio.SetButtonLamp(elevio.ButtonType(a.Button), a.Floor, true)
			}
		case a := <-drv_floors:
			if elev.Floor != a {
				elev.Floor = a
				if elevio.IsValidFloor(a) {
					elev.PrevValidFloor = a
					elevio.SetFloorIndicator(a)
				}
			}
		case a := <-drv_obstr:
			if a && elev.IsDoorOpen {
				elev.State = elevator.OBSTRUCTED
			}
			// Ignore for now
			//case a := <-drv_stop:
			//{
			//
			//}
		}
	}
}
