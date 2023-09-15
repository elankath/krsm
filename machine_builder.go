package krsm

import (
	"errors"
	"fmt"
	"slices"
)

type Builder[S State, E Event] struct {
	name               string
	states             []S
	edges              map[S][]edge[S, E]
	stateConfigurators map[S]*StateConfigurator[S, E]
	errors             []error
	parentStates       map[S]S
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

func (b *Builder[S, E]) ConfigureSubState(subState S, parentState S) *StateConfigurator[S, E] {
	subStateConfigurator := b.ConfigureState(subState)
	parentStates := b.parentStates
	if currParent, ok := parentStates[subState]; ok {
		if currParent != parentState {
			b.addError(ErrCannotHaveDiffParents, "cannot have parent %q as existing parent %q is already defined: %w", parentState, currParent)
		}
		return subStateConfigurator
	}
	b.parentStates[subState] = parentState
	return subStateConfigurator
}

func (c *StateConfigurator[S, E]) ConfigureState(state S) *StateConfigurator[S, E] {
	return c.builder.ConfigureState(state)
}

func (c *StateConfigurator[S, E]) ConfigureSubState(subState S, parentState S) *StateConfigurator[S, E] {
	return c.builder.ConfigureSubState(subState, parentState)
}

func (c *StateConfigurator[S, E]) Permit(event E, targetState S) *StateConfigurator[S, E] {
	edge := edge[S, E]{
		sourceState: c.state,
		event:       event,
		targetState: targetState,
	}
	stateEdges := c.builder.edges[c.state]
	if slices.Contains(stateEdges, edge) {
		c.builder.addError(ErrDuplicateEdge, "edge %q already defined: %w", edge)
		return c
	}
	stateEdges = append(stateEdges, edge)
	c.builder.edges[c.state] = stateEdges
	return c
}

// Build builds the Default State Machine. Internally ensures that indices are set on the edges for optimized
// adjacency list traversal
func (b *Builder[S, E]) Build() (StateMachine[S, E], error) {
	if len(b.errors) != 0 {
		return nil, errors.Join(b.errors...)
	}
	sm := defaultStateMachine[S, E]{
		name:         b.name,
		states:       b.states,
		edges:        b.edges,
		currentState: b.states[0],
	}
	return &sm, nil
}

func (b *Builder[S, E]) addError(sentinel error, format string, args ...any) {
	b.errors = append(b.errors, fmt.Errorf(format, args, sentinel))
}
