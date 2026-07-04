package fsm

import "testing"

func TestNextReturnsFollowingStage(t *testing.T) {
	if got := Next("proposal"); got != "critique" {
		t.Fatalf("Next() = %q, want critique", got)
	}
}

func TestIsStageRecognizesKnownAndUnknownStages(t *testing.T) {
	if !IsStage("review") {
		t.Fatal("IsStage(review) = false, want true")
	}
	if IsStage("unknown") {
		t.Fatal("IsStage(unknown) = true, want false")
	}
}

func TestCanTransitionAllowsForwardAndConfiguredBackSteps(t *testing.T) {
	if !CanTransition("intake", "proposal") {
		t.Fatal("CanTransition(intake, proposal) = false, want true")
	}
	if !CanTransition("review", "implementation") {
		t.Fatal("CanTransition(review, implementation) = false, want true")
	}
	if CanTransition("archived", "proposal") {
		t.Fatal("CanTransition(archived, proposal) = true, want false")
	}
}
