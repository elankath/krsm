package krsm

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestSimpleCatMachine(t *testing.T) {
	type CatState string
	type CatEvent string
	const (
		Purring    CatState = "Purring"
		Sleeping   CatState = "Sleeping"
		Scratching CatState = "Scratching"
		Biting     CatState = "Biting"

		Pet CatEvent = "PET"
		Hit CatEvent = "HIT"
	)
	var allStates = []CatState{Sleeping, Purring, Scratching, Biting}
	g := NewWithT(t)

	builder := NewBuilder[CatState, CatEvent]("CatMachine")
	catMachine, err := builder.ConfigureState(Sleeping).
		Target(Purring, Pet).
		Target(Scratching, Hit).
		ConfigureState(Purring).
		Target(Sleeping, Pet).
		Target(Biting, Hit).
		ConfigureState(Scratching).
		Target(Purring, Pet).
		Target(Biting, Hit).
		ConfigureState(Biting).
		Target(Scratching, Pet).
		Target(Biting, Hit).Build()

	g.Expect(err).To(BeNil())
	g.Expect(catMachine.States()).To(BeEquivalentTo(sets.New(allStates...)))

	g.Expect(catMachine.CurrentState()).To(BeEquivalentTo(Sleeping))
	transition, err := catMachine.Trigger(Pet, "So Fluffy!")
	g.Expect(err).To(BeNil())
	g.Expect(transition.TargetState).To(BeEquivalentTo(Purring))
	g.Expect(transition.SourceState).To(BeEquivalentTo(Sleeping))

}

func TestMachineWithSubStates(t *testing.T) {
	type DogState string
	type DogEvent string

	const (
		Asleep   DogState = "Asleep"
		Dreaming DogState = "Dreaming"
		Awake    DogState = "Awake"
		Barking  DogState = "Barking"
		Biting   DogState = "Biting"
		Wagging  DogState = "Wagging"
		Eating   DogState = "Eating"
		Dead     DogState = "Dead"

		Pet   DogEvent = "PET"
		Slap  DogEvent = "SLAP"
		Kick  DogEvent = "KICK"
		Feed  DogEvent = "FEED"
		Shoot DogEvent = "SHOOT"
	)
	g := NewWithT(t)
	//var allStates = []DogState{Asleep, Dreaming, Awake, Barking, Biting, Wagging, Eating}
	builder := NewBuilder[DogState, DogEvent]("DogMachine")
	dogMachine, err := builder.ConfigureState(Asleep).
		Target(Asleep, Pet).
		Target(Awake, Slap).
		Target(Barking, Kick).
		ConfigureState(Awake).
		Target(Biting, Slap).
		Target(Biting, Kick).
		Target(Eating, Feed).
		ConfigureSubState(Barking, Awake).
		Target(Biting, Slap, Kick).
		ConfigureSubState(Biting, Awake).
		Build()

	g.Expect(err).To(BeNil())
	g.Expect(dogMachine.CurrentState()).To(Equal(Asleep))
	transition, err := dogMachine.Trigger(Kick, "kicking Tommy")
	g.Expect(err).To(BeNil())
	g.Expect(dogMachine.CurrentState()).To(Equal(Barking))
	g.Expect(transition.SourceState).To(Equal(Asleep))
	g.Expect(transition.TargetState).To(Equal(Barking))
	g.Expect(transition.Event).To(Equal(Kick))
	transition, err = dogMachine.Trigger(Kick, "double-whammy")
	g.Expect(err).To(BeNil())
	g.Expect(dogMachine.CurrentState()).To(Equal(Biting))
	g.Expect(transition.SourceState).To(Equal(Barking))
	g.Expect(transition.TargetState).To(Equal(Biting))
	g.Expect(transition.Event).To(Equal(Kick))

	transition, err = dogMachine.Trigger(Feed, "dont bite, eat")
	g.Expect(err).To(BeNil())

	_, err = dogMachine.Trigger(Pet, "eat more")
	g.Expect(err).ToNot(BeNil())
	g.Expect(errors.Is(err, ErrCouldNotTransition)).To(BeTrue())

	//g.Expect(dogMachine.States()).To(BeEquivalentTo(sets.New(allStates...)))

}

func TestIllegalStateConfiguration(t *testing.T) {
	type HorseState string
	type HorseEvent string

	const (
		Asleep    HorseState = "Asleep"
		Awake     HorseState = "Awake"
		Galloping HorseState = "Galloping"
		Trotting  HorseState = "Trotting"

		Slap HorseEvent = "Slap"
		Kick HorseEvent = "Kick"
	)
	g := NewWithT(t)

	builder := NewBuilder[HorseState, HorseEvent]("HorseMachine")
	_, err := builder.ConfigureState(Asleep).
		Target(Awake, Slap).
		ConfigureState(Awake).
		ConfigureSubState(Awake, Asleep).
		Build()
	g.Expect(err).ToNot(BeNil())
	fmt.Println(err)
	g.Expect(errors.Is(err, ErrIllegalState)).To(BeTrue())
}
