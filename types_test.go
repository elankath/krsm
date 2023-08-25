package krsm

import (
	"testing"

	. "github.com/onsi/gomega"
)

type CatState string
type CatEvent string

func TestSimpleStateMachine(t *testing.T) {
	catStates := []CatState{"SLEEPING", "SCRATCHING", "PURRING", "BITING"}
	builder := NewBuilder[CatState, CatEvent](catStates[0])
	builder.ConfigureState("SLEEPING").
		Permit("pet", "PURRING").
		Permit("hit", "SCRATCHING").
		ConfigureState("SCRATCHING").
		Permit("hit", "BITING").
		Permit("pet", "PURRING").
		ConfigureState("BITING").
		Permit("pet", "SCRATCHING").
		Permit("hit", "BITING").
		ConfigureState("PURRING").
		Permit("pet", "SLEEPING").
		Permit("hit", "SCRATCHING")
	catm := builder.Build()

	g := NewWithT(t)
	g.Expect(catm.CurrentState()).To(BeEquivalentTo("SLEEPING"))

	nextState, err := catm.Trigger("pet")
	g.Expect(err).To(BeNil())
	g.Expect(nextState).To(BeEquivalentTo("PURRING"))
}
