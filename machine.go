package krsm

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DefaultStateMachine[S State, E Event] struct {
	currState  S
	stateEdges map[S][]edge[S, E]
}

func (m *DefaultStateMachine[S, E]) CurrentState() S {
	return m.currState
}

func (m *DefaultStateMachine[S, E]) Trigger(event E, message string) (transition Transition[S, E], err error) { //TODO: define custom error with error code
	// TODO, FIXME: Take Mutex
	edges := m.stateEdges[m.currState]
	//if !ok {
	//	return m.EmptyTransition, fmt.Errorf("%w: current state: %s has no edges ", ErrIllegalState, m.currState)
	//}
	for _, e := range edges {
		if e.SourceState != m.currState {
			err = fmt.Errorf("%w: current state: %s does not match source stage of edge: %v", ErrIllegalState, m.currState, e)
			return
		}
		if e.Event == event {
			targetState := e.TargetState
			transition = Transition[S, E]{
				CreatedTime: metav1.Now(),
				SourceState: e.SourceState,
				Event:       event,
				TargetState: e.TargetState,
				Message:     message,
			}
			m.currState = targetState
			return
		}
	}
	err = fmt.Errorf("no transition for Event: %s from SourceState: %s", event, m.currState)
	return
}
