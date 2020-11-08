package types

// Currently not used, could track on what level a site was visted
// Should be used in combination with bfs to avoid still having duplicate visits

type Tracker map[string]int

func NewTracker() *Tracker {
	tracker := Tracker(map[string]int{})
	return &tracker
}

func (tracker *Tracker) Add(s string, depth int) {
	// case not found -> set
	// case found smaller -> don't set
	// case found greater -> set
	v, found := (*tracker)[s]
	if !found || v > depth {
		(*tracker)[s] = depth
	}
}

func (tracker *Tracker) ShouldVisit(s string, depth int) bool {
	v, found := (*tracker)[s]
	return !found || depth >= v
}
