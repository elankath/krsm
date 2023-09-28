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
	childToParentStates map[S]S // DISCUSS THIS: proposal-1.
}

func (m *defaultStateMachine[S, E]) Name() string {
	return m.name
}

func (m *defaultStateMachine[S, E]) States() sets.Set[S] {
	return sets.New(m.states...)
}

func (m *defaultStateMachine[S, E]) Trigger(triggerEvent E, resource Resource[S, E], message string) (transition Transition[S, E], err error) {
	sourceState := resource.CurrentState()
	stateEdges := m.edges[sourceState]
	for {
		for _, e := range stateEdges {
			if e.event == triggerEvent {
				if e.guard != nil && !e.guard(resource) {
					continue
				}
				transition = Transition[S, E]{
					CreatedTime: metav1.Now(),
					SourceState: resource.CurrentState(),
					Event:       triggerEvent,
					TargetState: e.targetState,
					Message:     message,
				}
				resource.SetTransition(transition)
				return
			}
		}
		parentState, ok := m.childToParentStates[sourceState]
		if !ok {
			err = fmt.Errorf("no edge for event %q from resource state %q: %w", triggerEvent, resource.CurrentState(), ErrCouldNotTransition)
			return
		}
		sourceState = parentState
		stateEdges = m.edges[parentState]
	}
}
