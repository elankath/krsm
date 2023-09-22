package krsm

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type defaultStateMachine[S State, E Event] struct {
	name   string
	states []S                // TODO: Change to [][]S when sub-states are introduced or have separate childToParentStates map
	edges  map[S][]edge[S, E] // adjacency map representation,
	// ALTERNATIVELY: use adjacency matrix ie [][]struct{} or [][]int and a []S stateIndices and []Edge[S, E] edgeIndices
	// If there is an edge from vertex i to j, mark adjMat[i][j] as 1.
	// If there is no edge from vertex i to j, mark adjMat[i][j] as 0.
	currentState        S
	childToParentStates map[S]S // DISCUSS THIS: proposal-1.
}

func (m *defaultStateMachine[S, E]) Name() string {
	return m.name
}

func (m *defaultStateMachine[S, E]) CurrentState() S {
	return m.currentState
}

func (m *defaultStateMachine[S, E]) States() sets.Set[S] {
	return sets.New(m.states...)
}

func (m *defaultStateMachine[S, E]) Trigger(triggerEvent E, message string) (transition Transition[S, E], err error) { //TODO: define custom error with error code
	// TODO, FIXME: Take Mutex and then call Trigger mutex
	stateEdges := m.edges[m.currentState]
	// TODO: When adding sub-states, I must create a slice of all state edges from the parent state also.
	// This is easy to do with using
	//  parentState := childToParentStates map[S]S
	//  parentEdges := m.edges[parentState] and then check transitions on the parent state.
	//  We repeat this in a loop until we have found a transition or parentState is nil.
	sourceState := m.CurrentState()
	for {
		for _, e := range stateEdges {
			if e.sourceState != sourceState {
				err = fmt.Errorf("current state %q does not match source stage of event %v: %w", m.currentState, e, ErrIllegalState)
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
		parentState, ok := m.childToParentStates[sourceState]
		if !ok {
			err = fmt.Errorf("no edge for event %q from source state %q: %w", triggerEvent, m.currentState, ErrCouldNotTransition)
			return
		}
		sourceState = parentState
		stateEdges = m.edges[parentState]
	}
}
