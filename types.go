package krsm

import (
	"errors"
	"fmt"

	"golang.org/x/exp/constraints"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	ErrUnknownState = errors.New("invalid state")

	ErrInvalidTransition = errors.New("invalid transition")

	// ErrIllegalState is a sentinel error indicating that the state machine is in an illegal state.
	ErrIllegalState = errors.New("illegal state")

	// ErrDuplicateEdge is a sentinel error indicating that a duplicate edge has been defined by the consumer.
	ErrDuplicateEdge = errors.New("duplicate edge")

	// ErrCouldNotTransition is a sentinel error indicating that one could not transition from the current state
	ErrCouldNotTransition = errors.New("could not transition from current state")

	// ErrCannotHaveDiffParents is a sentinel error indicating that a state cannot have different parents
	ErrCannotHaveDiffParents = errors.New("state cannot have diff parent")

	// ErrNoOutEdges is a sentinel error indicating that a state has no outgoing edges
	ErrNoOutEdges = errors.New("no out edges")
)

type State interface {
	~string
	constraints.Ordered
}

type StateConfigurator[S State, E Event] interface {
	ConfigureState(state S) StateConfigurator[S, E]
	ConfigureSubState(subState S, parentState S) StateConfigurator[S, E]
	Target(targetState S, events ...E) StateConfigurator[S, E]
	Build() (StateMachine[S, E], error)
}

type Event interface {
	~string
	constraints.Ordered // if it is not used in a map then remove this
}

type Transition[S State, E Event] struct {
	CreatedTime metav1.Time
	SourceState S
	Event       E
	TargetState S
	Message     string
}

type GuardPredicate func() (bool, error) // TODO: define this carefully later

// TODO define this later.
type Guard[S State] struct {
}

type StateMachine[S State, E Event] interface {
	Name() string
	CurrentState() S
	Trigger(event E, message string) (transition Transition[S, E], err error)
	States() sets.Set[S]
}

type edge[S State, E Event] struct {
	event       E
	sourceState S
	targetState S
	// TODO: Add Guard
}

func (e *edge[S, E]) String() string {
	return fmt.Sprintf("(%s-%s-%s)", e.sourceState, e.event, e.targetState)
}
