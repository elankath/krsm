package krsm

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestSimpleCatMachine(t *testing.T) {
	var allStates = []CatState{CatStates.Sleeping, CatStates.Purring, CatStates.Scratching, CatStates.Biting}
	g := NewWithT(t)

	builder := NewBuilder[CatState, CatEvent]("CatMachine")
	catMachine, err := builder.ConfigureState(CatStates.Sleeping).
		Target(CatStates.Purring, CatEvents.Pet).
		Target(CatStates.Scratching, CatEvents.Hit).
		ConfigureState(CatStates.Purring).
		Target(CatStates.Sleeping, CatEvents.Pet).
		Target(CatStates.Biting, CatEvents.Hit).
		ConfigureState(CatStates.Scratching).
		Target(CatStates.Purring, CatEvents.Pet).
		Target(CatStates.Biting, CatEvents.Hit).
		ConfigureState(CatStates.Biting).
		Target(CatStates.Scratching, CatEvents.Pet).
		Target(CatStates.Biting, CatEvents.Hit).Build()

	g.Expect(err).To(BeNil())
	g.Expect(catMachine.States()).To(BeEquivalentTo(sets.New(allStates...)))

	cat := Cat[CatState, CatEvent]{name: "sammy", namespace: "bangalore", currentState: CatStates.Sleeping}
	transition, err := catMachine.Trigger(CatEvents.Pet, cat, "So Fluffy!")
	g.Expect(err).To(BeNil())
	g.Expect(transition.TargetState).To(BeEquivalentTo(CatStates.Purring))
	g.Expect(transition.SourceState).To(BeEquivalentTo(CatStates.Sleeping))

	//TODO: add more sub-tests for transitions.

}

func TestMachineWithSubStates(t *testing.T) {
	g := NewWithT(t)
	//var allStates = []DogState{Asleep, Dreaming, Awake, Barking, Biting, Wagging, Eating}
	builder := NewBuilder[DogState, DogEvent]("DogMachine")
	dogMachine, err := builder.ConfigureState(DogStates.Asleep).
		Target(DogStates.Asleep, DogEvents.Pet).
		Target(DogStates.Awake, DogEvents.Slap).
		Target(DogStates.Barking, DogEvents.Kick).
		ConfigureState(DogStates.Awake).
		Target(DogStates.Biting, DogEvents.Slap).
		Target(DogStates.Biting, DogEvents.Kick).
		Target(DogStates.Eating, DogEvents.Feed).
		ConfigureSubState(DogStates.Barking, DogStates.Awake).
		Target(DogStates.Biting, DogEvents.Slap, DogEvents.Kick).
		ConfigureSubState(DogStates.Biting, DogStates.Awake).
		Build()

	myDog := &Dog[DogState, DogEvent]{name: "tommy", namespace: "pet", currentState: DogStates.Asleep}
	g.Expect(err).To(BeNil())
	g.Expect(myDog.CurrentState()).To(Equal(DogStates.Asleep))

	//TODO: Make the below as sub-tests with right label
	transition, err := dogMachine.Trigger(DogEvents.Kick, myDog, "kicking Tommy")
	g.Expect(err).To(BeNil())
	g.Expect(myDog.CurrentState()).To(Equal(DogStates.Barking))
	g.Expect(transition.SourceState).To(Equal(DogStates.Asleep))
	g.Expect(transition.TargetState).To(Equal(DogStates.Barking))
	g.Expect(transition.Event).To(Equal(DogEvents.Kick))

	transition, err = dogMachine.Trigger(DogEvents.Kick, myDog, "double-whammy")
	g.Expect(err).To(BeNil())
	g.Expect(myDog.CurrentState()).To(Equal(DogStates.Biting))
	g.Expect(transition.SourceState).To(Equal(DogStates.Barking))
	g.Expect(transition.TargetState).To(Equal(DogStates.Biting))
	g.Expect(transition.Event).To(Equal(DogEvents.Kick))

	transition, err = dogMachine.Trigger(DogEvents.Feed, myDog, "dont bite, eat")
	g.Expect(err).To(BeNil())

	_, err = dogMachine.Trigger(DogEvents.Pet, myDog, "pet more")
	g.Expect(err).ToNot(BeNil())
	g.Expect(errors.Is(err, ErrCouldNotTransition)).To(BeTrue())

	//g.Expect(dogMachine.States()).To(BeEquivalentTo(sets.New(allStates...)))

}

func TestIllegalStateConfiguration(t *testing.T) {
	type HorseState string
	type HorseEvent string

	const (
		Asleep HorseState = "Asleep"
		Awake  HorseState = "Awake"

		Slap HorseEvent = "Slap"
	)
	g := NewWithT(t)

	builder := NewBuilder[HorseState, HorseEvent]("HorseMachine")
	_, err := builder.ConfigureState(Asleep).
		Target(Awake, Slap).
		ConfigureState(Awake).
		ConfigureSubState(Awake, Asleep).
		Build()
	g.Expect(err).ToNot(BeNil())
	g.Expect(errors.Is(err, ErrIllegalState)).To(BeTrue())
}

//func TestMachineWithGuards(t *testing.T) {
//	type DogState string
//	type DogEvent string
//	type Dog struct {
//		name string
//		age  int
//	}
//	puppy := Dog{
//		name: "puppy",
//		age:  1,
//	}
//	granny := Dog{
//		name: "granny",
//		age:  10,
//	}
//
//	const (
//		Barking DogState = "Barking"
//		Biting  DogState = "Biting"
//		Wagging DogState = "Wagging"
//
//		Pet  DogEvent = "PET"
//		Slap DogEvent = "SLAP"
//		Kick DogEvent = "KICK"
//	)
//
//	g := NewWithT(t)
//	builder := NewBuilder[DogState, DogEvent]("DogMachine")
//	dogMachine, err := builder.ConfigureState(Barking).
//		Target(Wagging, Pet).
//		TargetWithGuard(Biting, Slap).
//		Target(Barking, Kick).
//		Build()
//}

type CatState string
type CatEvent string

var CatStates = struct {
	Purring    CatState
	Sleeping   CatState
	Scratching CatState
	Biting     CatState
	Pet        CatEvent
	Hit        CatEvent
}{
	Purring:    "Purring",
	Sleeping:   "Sleeping",
	Scratching: "Scratching",
}
var CatEvents = struct {
	Pet CatEvent
	Hit CatEvent
}{
	Pet: "Pet",
	Hit: "hit",
}

type Cat[S CatState, E CatEvent] struct {
	name           string
	namespace      string
	currentState   S
	lastTransition Transition[S, E]
}

func (c Cat[S, E]) GetNamespace() string {
	return c.namespace
}

func (c Cat[S, E]) GetName() string {
	return c.name
}

func (c Cat[S, E]) CurrentState() S {
	return c.currentState
}

func (c Cat[S, E]) SetTransition(transition Transition[S, E]) {
	c.lastTransition = transition
	c.currentState = transition.TargetState
}

type DogState string
type DogEvent string

var DogStates = struct {
	Asleep  DogState
	Awake   DogState
	Barking DogState
	Biting  DogState
	Eating  DogState
}{
	Asleep:  "Asleep",
	Awake:   "Awake",
	Barking: "Barking",
	Biting:  "Biting",
	Eating:  "Eating",
}
var DogEvents = struct {
	Pet  DogEvent
	Slap DogEvent
	Kick DogEvent
	Feed DogEvent
}{
	Pet:  "PET",
	Slap: "SLAP",
	Kick: "KICK",
	Feed: "FEED",
}

type Dog[S DogState, E DogEvent] struct {
	name           string
	namespace      string
	currentState   S
	lastTransition Transition[S, E]
}

func (c *Dog[S, E]) GetNamespace() string {
	return c.namespace
}

func (c *Dog[S, E]) GetName() string {
	return c.name
}

func (c *Dog[S, E]) CurrentState() S {
	return c.currentState
}

func (c *Dog[S, E]) SetTransition(transition Transition[S, E]) {
	c.lastTransition = transition
	c.currentState = transition.TargetState
}
