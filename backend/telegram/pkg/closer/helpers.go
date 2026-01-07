package closer

import "slices"

// sortClosersByOrder sorting closers
func (m *Manager) sortClosersByOrder() []Closer {
	orderIndex := make(map[string]int)
	for i, name := range m.shutdownOrder {
		orderIndex[name] = i
	}

	sorted := make([]Closer, len(m.closers))
	copy(sorted, m.closers)

	slices.SortFunc(sorted, func(a, b Closer) int {
		idxA, okA := orderIndex[a.Name()]
		idxB, okB := orderIndex[b.Name()]

		if okA && okB {
			return idxA - idxB
		}
		if okA {
			return -1
		}
		if okB {
			return 1
		}
		return 0
	})

	return sorted
}
