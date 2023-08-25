package krsm

import (
	"slices"
)

type Builder[S State, E Event] struct {
	name               string
	states             []S
	edges              map[S][]edge[S, E]
	stateConfigurators map[S]*StateConfigurator[S, E]
}

type StateConfigurator[S State, E Event] struct {
	builder *Builder[S, E]
	state   S
}

func NewBuilder[S State, E Event](name string) *Builder[S, E] {
	return &Builder[S, E]{
		name:               name,
		stateConfigurators: make(map[S]*StateConfigurator[S, E]),
		edges:              make(map[S][]edge[S, E]),
	}
}

func (b *Builder[S, E]) ConfigureState(state S) *StateConfigurator[S, E] {
	stateConfigurator, ok := b.stateConfigurators[state]
	if ok {
		return stateConfigurator
	}
	stateConfigurator = &StateConfigurator[S, E]{
		builder: b,
		state:   state,
	}
	b.states = append(b.states, state)
	b.stateConfigurators[state] = stateConfigurator
	return stateConfigurator
}

func (c *StateConfigurator[S, E]) ConfigureState(state S) *StateConfigurator[S, E] {
	return c.builder.ConfigureState(state)
}

func (c *StateConfigurator[S, E]) Permit(event E, targetState S) *StateConfigurator[S, E] {
	edge := edge[S, E]{
		sourceState: c.state,
		event:       event,
		targetState: targetState,
	}
	stateEdges := c.builder.edges[c.state]
	if slices.Contains(stateEdges, edge) { // design defect in code
		panic("already defined edge: " + edge.String())
	}
	stateEdges = append(stateEdges, edge)
	c.builder.edges[c.state] = stateEdges
	return c
}

// Build builds the Default State Machine. Internally ensures that indices are set on the edges for optimized
// adjacency list traversal
func (b *Builder[S, E]) Build() *DefaultStateMachine[S, E] {
	sm := DefaultStateMachine[S, E]{
		name:         b.name,
		states:       b.states,
		edges:        b.edges,
		currentState: b.states[0],
	}
	return &sm
}
