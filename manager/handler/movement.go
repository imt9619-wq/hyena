package handler

import (
	"reflect"
	"slices"
	"sync/atomic"
)

type doMovement interface{
	// doAction simply add a force to the player, for example if a player is on jumping 
	// action, a Y force will be added if onGround is true if only gravity is acting 
	// on the player, the Y force will be 0 instead of the force of the gravity
	doAction(*playerMovement) 
}



// playerMovement is the force and momentnum applied onto the client, 
// currentAction funcs is going to use the playerMovement and manlapulate it,
// then a playerAuthInput packet will be sent after the force and other field is applied
// and return the player next position to be used in the playerAuthInput packet.
type playerMovement struct {
	sc *sessionConf
	currentAction []doMovement
}


func (pm *playerMovement) tick(){
	defer pm.sc.playerState.Unlock()
	pm.sc.playerState.Lock()
	if !pm.sc.playerState.onReset.CompareAndSwap(true, false){
		for _, aMove := range pm.currentAction{
			aMove.doAction(pm)
		}
	}
	pm.applyForceOnState()
}


// the phycsis caluation is done on this function, player coordaniate 
// is changed in here then will get writen into playerAuthInput
func (pm *playerMovement) applyForceOnState(){

}


// return the index of the action and true in the pm.currentAction if exist else -1 and false
func (pm *playerMovement) getAction(dm doMovement) (int, bool) {
	targetType := reflect.TypeOf(dm)
	for i, act := range pm.currentAction {
		if reflect.TypeOf(act) == targetType {
			return i, true
		}
	}
	return -1, false
}

func (pm *playerMovement) AddAction(dm doMovement) {
	_, ok := pm.getAction(dm)
	if ok{
		return
	}
	pm.currentAction = append(pm.currentAction, dm)
}


func (pm *playerMovement) RemoveAction(dm doMovement){
	index , ok := pm.getAction(dm)
	if !ok{
		return
	}
	pm.currentAction = slices.Delete(pm.currentAction, index, index+1)
}


func newPlayerMovement(sc *sessionConf) *playerMovement{
	onGround := &atomic.Bool{}
	onGround.Store(true)
	
	return &playerMovement{
		sc: sc,
		currentAction: make([]doMovement, 0, 10),
	}
}
