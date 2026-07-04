package fsm

var Stages = []string{"intake", "context", "proposal", "critique", "decision", "implementation", "review", "verification", "handoff", "archived"}

func Next(stage string) string {
	for i, s := range Stages {
		if s == stage && i+1 < len(Stages) {
			return Stages[i+1]
		}
	}
	return ""
}

func IsStage(stage string) bool {
	return index(stage) >= 0
}

func CanTransition(from, to string) bool {
	if from == "archived" {
		return false
	}
	fromI, toI := index(from), index(to)
	if fromI < 0 || toI < 0 {
		return false
	}
	// Forward skips are allowed because blackboard supports simplified workflows.
	if toI > fromI {
		return true
	}
	back := map[string][]string{
		"review":       {"implementation", "proposal", "decision"},
		"verification": {"implementation", "review"},
		"critique":     {"proposal"},
		"decision":     {"proposal", "critique"},
	}
	for _, x := range back[from] {
		if x == to {
			return true
		}
	}
	return false
}

func index(stage string) int {
	for i, s := range Stages {
		if s == stage {
			return i
		}
	}
	return -1
}
