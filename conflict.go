package pd

import (
	"log"
	"sort"
	"sync"
)

// Conflict represents an overlap in two RenderedScheduleEntries, where the
// Left side begins first, but the Right side begins before the Left side ends.
//
// Currently, we only support finding conflicts grouped to a particular user ID,
// so callers can assume that Left.User.ID == Right.User.ID and use either. This
// may have to change to support other modes of conflict detection.
type Conflict struct {
	Left  *RenderedScheduleEntry
	Right *RenderedScheduleEntry
}

// FindConflictsByUser returns a mapping of user ID to a list of Conflicts for
// that user in the given list of RenderedScheduleEntries.
//
// It also logs any conflicts found as a side-effect, which may change in the
// future.
func FindConflictsByUser(entries []*RenderedScheduleEntry) map[string][]*Conflict {
	entriesByUser := RenderedScheduleEntries(entries).GroupBy(func(entry *RenderedScheduleEntry) string {
		return entry.User.ID
	})

	var (
		m       sync.Mutex
		wg      sync.WaitGroup
		results = make(map[string][]*Conflict, len(entriesByUser))
	)

	for userID, entries := range entriesByUser {
		wg.Add(1)

		go func(userID string, entries []*RenderedScheduleEntry) {
			defer wg.Done()

			conflicts := []*Conflict{}

			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Start.Before(entries[j].Start)
			})

			for i, left := range entries {
				for j := i + 1; j < len(entries); j++ {
					right := entries[j]

					if !right.Start.Before(left.End) { // if left.End <= right.Start
						// All good, RHS doesn't start until at least after LHS
						// ends. Stop scanning for conflicts related to LHS.
						break
					}

					log.Printf("CONFLICT: %s is in both %q and %q from %s to %s\n", left.User.Summary, left.Schedule, right.Schedule, right.Start, left.End)

					conflicts = append(conflicts, &Conflict{Left: left, Right: right})
				}
			}

			m.Lock()
			defer m.Unlock()

			results[userID] = conflicts
		}(userID, entries)
	}

	wg.Wait()

	return results
}
