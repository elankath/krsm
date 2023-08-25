package krsm

import "k8s.io/apimachinery/pkg/util/sets"

type Builder[S State, E Event] struct {
	initialState S
	states       sets.Set[S]
	edges        sets.Set[edge[S, E]]
}

type StateConfigurator[S State, E Event] struct {
	builder *Builder[S, E]
	state   S
}

func NewBuilder[S State, E Event](initialState S) *Builder[S, E] {
	return &Builder[S, E]{
		initialState: initialState,
		states:       sets.New[S](initialState),
		edges:        sets.New[edge[S, E]](),
	}
}

func (b *Builder[S, E]) ConfigureState(state S) *StateConfigurator[S, E] {
	return &StateConfigurator[S, E]{
		builder: b,
		state:   state,
	}
}

func (b *Builder[S, E]) Build() *DefaultStateMachine[S, E] {
	var stateEdges = make(map[S][]edge[S, E], len(b.states))
	for s := range b.states {
		stateEdges[s] = make([]edge[S, E], 0)
	}
	for e := range b.edges {
		stateEdges[e.SourceState] = append(stateEdges[e.SourceState], e)
	}
	sm := DefaultStateMachine[S, E]{
		currState:  b.initialState,
		stateEdges: stateEdges,
	}
	return &sm
}

func (b *Builder[S, E]) addEdge(edge edge[S, E]) {
	b.edges.Insert(edge)
}

func (c *StateConfigurator[S, E]) Permit(event E, targetState S) *StateConfigurator[S, E] {
	edge := edge[S, E]{
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
