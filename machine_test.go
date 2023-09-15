package krsm

import (
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

	builder := NewBuilder[CatState, CatEvent]("CatMachine")
	builder.ConfigureState(Sleeping).
		Permit(Pet, Purring).
		Permit(Hit, Scratching).
		ConfigureState(Purring).
		Permit(Pet, Sleeping).
		Permit(Hit, Biting).
		ConfigureState(Scratching).
		Permit(Pet, Purring).
		Permit(Hit, Biting).
		ConfigureState(Biting).
		Permit(Pet, Scratching).
		Permit(Hit, Biting)

	catMachine, err := builder.Build()
	g := NewWithT(t)
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

		Pet  DogEvent = "PET"
		Slap DogEvent = "SLAP"
		Kick DogEvent = "KICK"
		Feed DogEvent = "FEED"
	)
	var allStates = []DogState{Asleep, Dreaming, Awake, Barking, Biting, Wagging, Eating}
	builder := NewBuilder[DogState, DogEvent]("DogMachine")
	builder.ConfigureState(Asleep).
		Permit(Pet, Asleep).
		Permit(Slap, Awake).
		Permit(Kick, Barking).
		//ConfigureSubState()

}
