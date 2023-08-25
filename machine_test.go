package krsm

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestSimpleCatMachine(t *testing.T) {
	type CatState string
	type CatEvent string
	var (
		Purring    CatState = "Purring"
		Sleeping   CatState = "Sleeping"
		Scratching CatState = "Scratching"
		Biting     CatState = "Biting"

		Pet CatEvent = "PET"
		Hit CatEvent = "HIT"
	)
	builder := NewBuilder[CatState, CatEvent](Sleeping)
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

	var catMachine StateMachine[CatState, CatEvent] = builder.Build()

	g := NewWithT(t)
	g.Expect(catMachine.CurrentState()).To(BeEquivalentTo("SLEEPING"))

	transition, err := catMachine.Trigger("pet", "So Fluffy!")
	g.Expect(err).To(BeNil())
	g.Expect(transition.TargetState).To(BeEquivalentTo("PURRING"))
}
