package krsm

import (
	"errors"
	"fmt"
	"slices"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/sets"
)

type Builder[S State, E Event] struct {
	name                string
	states              []S
	edges               map[S][]edge[S, E]
	stateConfigurators  map[S]*defaultStateConfigurator[S, E]
	errors              []error
	childToParentStates map[S]S
}

type defaultStateConfigurator[S State, E Event] struct {
	builder *Builder[S, E]
	state   S
}

func NewBuilder[S State, E Event](name string) *Builder[S, E] {
	return &Builder[S, E]{
		name:                name,
		stateConfigurators:  make(map[S]*defaultStateConfigurator[S, E]),
		edges:               make(map[S][]edge[S, E]),
		childToParentStates: make(map[S]S),
	}
}

func (b *Builder[S, E]) newStateConfigurator(state S) *defaultStateConfigurator[S, E] {
	stateConfigurator, ok := b.stateConfigurators[state]
	if ok {
		return stateConfigurator
	}
	stateConfigurator = &defaultStateConfigurator[S, E]{
		builder: b,
		state:   state,
	}
	b.states = append(b.states, state)
	b.stateConfigurators[state] = stateConfigurator
	return stateConfigurator
}

func (b *Builder[S, E]) ConfigureSubState(subState S, parentState S) StateConfigurator[S, E] {
	subStateConfigurator := b.newStateConfigurator(subState)
	if b.getParentStates().Has(subState) {
		b.addError(ErrIllegalState, "sub-state %q is already defined as top level state: %w", subState)
		return &noopStateConfigurator[S, E]{builder: b}
	}
	parentStates := b.childToParentStates
	if currParent, ok := parentStates[subState]; ok {
		if currParent != parentState {
			b.addError(ErrCannotHaveDiffParents, "cannot have parent %q as existing parent %q is already defined: %w", parentState, currParent)
		}
		return subStateConfigurator
	}
	b.childToParentStates[subState] = parentState
	return subStateConfigurator
}

func (b *Builder[S, E]) ConfigureState(state S) StateConfigurator[S, E] {
	if _, ok := b.childToParentStates[state]; ok {
		b.addError(ErrIllegalState, "state %q is already a sub-state: %w", state)
		return &noopStateConfigurator[S, E]{builder: b}
	}
	return b.newStateConfigurator(state)
}

func (c *defaultStateConfigurator[S, E]) ConfigureState(state S) StateConfigurator[S, E] {
	return c.builder.newStateConfigurator(state)
}

func (c *defaultStateConfigurator[S, E]) ConfigureSubState(subState S, parentState S) StateConfigurator[S, E] {
	return c.builder.ConfigureSubState(subState, parentState)
}

func (c *defaultStateConfigurator[S, E]) Build() (StateMachine[S, E], error) {
	return c.builder.Build()
}

func (c *defaultStateConfigurator[S, E]) Target(targetState S, events ...E) StateConfigurator[S, E] {
	for _, event := range events {
		e := edge[S, E]{
			sourceState: c.state,
			event:       event,
			targetState: targetState,
		}
		c.addEdge(e)
	}
	return c
}

func (c *defaultStateConfigurator[S, E]) TargetWithGuard(targetState S, event E, guardLabel string, guard Guard[S, E]) StateConfigurator[S, E] {
	e := edge[S, E]{
		sourceState: c.state,
		event:       event,
		targetState: targetState,
		guardLabel:  guardLabel,
		guard:       guard,
	}
	c.addEdge(e)
	return c
}

func (c *defaultStateConfigurator[S, E]) addEdge(e edge[S, E]) {
	stateEdges := c.builder.edges[c.state]
	if slices.ContainsFunc(stateEdges, func(k edge[S, E]) bool {
		if k.event == e.event && k.sourceState == e.sourceState && k.targetState == e.targetState {
			return true
		}
		return false
	}) {
		c.builder.addError(ErrDuplicateEdge, "edge %q already defined: %w", e)
		return
	}
	stateEdges = append(stateEdges, e)
	c.builder.edges[c.state] = stateEdges
}

// Build builds the Default State Machine. Internally ensures that indices are set on the edges for optimized
// adjacency list traversal
func (b *Builder[S, E]) Build() (stateMachine StateMachine[S, E], err error) {
	initialState := b.states[0]
	b.validateInitialState(initialState)
	if len(b.errors) != 0 {
		err = errors.Join(b.errors...)
		return
	}
	stateMachine = &defaultStateMachine[S, E]{
		name:                b.name,
		states:              b.states,
		edges:               b.edges,
		childToParentStates: b.childToParentStates,
	}
	return
}

func (b *Builder[S, E]) getParentStates() sets.Set[S] {
	parentStates := sets.New(b.states...)
	parentStates.Delete(maps.Keys(b.childToParentStates)...)
	return parentStates
}

func (b *Builder[S, E]) validateInitialState(initialState S) {
	if len(b.edges[initialState]) == 0 {
		b.addError(ErrNoOutEdges, "no outgoing edges from %q: %w", initialState)
	}
}

func (b *Builder[S, E]) addError(sentinel error, format string, args ...any) {
	b.errors = append(b.errors, fmt.Errorf(format, args, sentinel))
}

type noopStateConfigurator[S State, E Event] struct {
	builder *Builder[S, E]
}

func (n *noopStateConfigurator[S, E]) TargetWithGuard(_ S, _ E, _ string, _ Guard[S, E]) StateConfigurator[S, E] {
	return n
}

func (n *noopStateConfigurator[S, E]) ConfigureState(_ S) StateConfigurator[S, E] {
	return n
}

func (n *noopStateConfigurator[S, E]) ConfigureSubState(_ S, _ S) StateConfigurator[S, E] {
	return n
}

func (n *noopStateConfigurator[S, E]) Target(_ S, _ ...E) StateConfigurator[S, E] {
	return n
}

func (n *noopStateConfigurator[S, E]) Build() (StateMachine[S, E], error) {
	return nil, errors.Join(n.builder.errors...)
}
