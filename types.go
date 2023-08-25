package krsm

import (
	"fmt"

	"golang.org/x/exp/constraints"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

/*
	state
	event
	transition
	sub-state
	guard
	state-machine
*/

type State interface {
	~string
	constraints.Ordered
}

type Event interface {
	~string
	constraints.Ordered // if it is not used in a map then remove this
}

type Edge[S State, E Event] struct {
	SourceState S
	Event       E
	TargetState S
}

type Transition[S State, E Event] struct {
	edge        Edge[S, E]
	createdTime metav1.Time
	message     string
}

type GuardPredicate func() (bool, error) // TODO: define this carefully later

// TODO define this later.
type Guard[S State] struct {
}

//type Node[S State] struct {
//	state S
//	// TODO could have metadata represented as a map
//}

type StateMachine[S State, E Event] struct {
	currState S
	states    sets.Set[S]
	edges     sets.Set[Edge[S, E]]
}

func (m *StateMachine[S, E]) CurrentState() S {
	return m.currState
}

func (m *StateMachine[S, E]) Trigger(event E) (targetState S, err error) { //TODO: define custom error with error code
	for edge := range m.edges {
		if edge.SourceState == m.currState && edge.Event == event {
			// TODO: perform the guard check here
			targetState = edge.TargetState
			m.currState = targetState
			return
		}
	}
	err = fmt.Errorf("no transition for Event: %s from SourceState: %s", event, m.currState) //TODO: use error code
	return
}

type Builder[S State, E Event] struct {
	initialState S
	states       sets.Set[S]
	edges        sets.Set[Edge[S, E]]
}

type StateConfigurator[S State, E Event] struct {
	builder *Builder[S, E]
	state   S
}

func NewBuilder[S State, E Event](initialState S) *Builder[S, E] {
	return &Builder[S, E]{
		initialState: initialState,
		states:       sets.New[S](initialState),
		edges:        sets.New[Edge[S, E]](),
	}
}

func (b *Builder[S, E]) ConfigureState(state S) *StateConfigurator[S, E] {
	return &StateConfigurator[S, E]{
		builder: b,
		state:   state,
	}
}

func (b *Builder[S, E]) Build() StateMachine[S, E] {
	sm := StateMachine[S, E]{
		currState: b.initialState,
		states:    b.states,
		edges:     b.edges,
	}
	return sm
}

func (b *Builder[S, E]) addEdge(edge Edge[S, E]) {
	b.edges.Insert(edge)
}

func (c *StateConfigurator[S, E]) Permit(event E, targetState S) *StateConfigurator[S, E] {
	edge := Edge[S, E]{
		SourceState: c.state,
		Event:       event,
		TargetState: targetState,
	}
	c.builder.addEdge(edge)
	c.builder.states.Insert(targetState)
	return c
}

func (c *StateConfigurator[S, E]) ConfigureState(state S) *StateConfigurator[S, E] {
	return &StateConfigurator[S, E]{
		builder: c.builder,
		state:   state,
	}
}

/*
	machineBuilder = NewBuilder()

	machineBuilder.ConfigureState()
		.permit()
		.permit()

	builder.

	builder.ConfigureState(stateName)
		.permit(event, targetState)
*/
