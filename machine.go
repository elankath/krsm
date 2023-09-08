package krsm

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type DefaultStateMachine[S State, E Event] struct {
	name   string
	states []S                // TODO: Change to [][]S when sub-states are introduced
	edges  map[S][]edge[S, E] // adjacency map representation
	// ALTERNATIVELY: use adjacency matrix ie [][]struct{} or [][]int and a []S stateIndices and []Edge[S, E] edgeIndices
	// If there is an edge from vertex i to j, mark adjMat[i][j] as 1.
	// If there is no edge from vertex i to j, mark adjMat[i][j] as 0.
	currentState S
}

func (m *DefaultStateMachine[S, E]) Name() string {
	return m.name
}

func (m *DefaultStateMachine[S, E]) CurrentState() S {
	return m.currentState
}

func (m *DefaultStateMachine[S, E]) States() []S {
	return m.states
}

func (m *DefaultStateMachine[S, E]) StatesSet() sets.Set[S] {
	return sets.New(m.states...)
}

func (m *DefaultStateMachine[S, E]) Trigger(triggerEvent E, message string) (transition Transition[S, E], err error) { //TODO: define custom error with error code
	//// TODO, FIXME: Take Mutex and then call Trigger mutex
	stateEdges := m.edges[m.currentState]
	for _, e := range stateEdges {
		if e.sourceState != m.currentState {
			err = fmt.Errorf("%w: current state %q does not match source stage of e %v", ErrIllegalState, m.currentState, e)
			return
		}
		if e.event == triggerEvent {
			transition = Transition[S, E]{
				CreatedTime: metav1.Now(),
				SourceState: m.currentState,
				Event:       triggerEvent,
				TargetState: e.targetState,
				Message:     message,
			}
			m.currentState = transition.TargetState
			return
		}
	}
	err = fmt.Errorf("no transition for Event: %s from SourceState: %s", triggerEvent, m.currentState)
	return
}
