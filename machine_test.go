package krsm

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestSimpleCatMachine(t *testing.T) {
	type CatState string
	type CatEvent string
	var (
		Purring    CatState = "Purring"
		Sleeping   CatState = "Sleeping"
		Scratching CatState = "Scratching"
		Biting     CatState = "Biting"

		allStates = []CatState{Sleeping, Purring, Scratching, Biting}

		Pet CatEvent = "PET"
		Hit CatEvent = "HIT"
	)
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

	var catMachine = builder.Build()
	g := NewWithT(t)
	g.Expect(catMachine.StatesSet()).To(BeEquivalentTo(sets.New(allStates...)))

	g.Expect(catMachine.currentState).To(BeEquivalentTo(Sleeping))
	transition, err := catMachine.Trigger(Pet, "So Fluffy!")
	g.Expect(err).To(BeNil())
	g.Expect(transition.TargetState).To(BeEquivalentTo(Purring))
	g.Expect(transition.SourceState).To(BeEquivalentTo(Sleeping))

}
