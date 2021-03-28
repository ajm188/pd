package pd

import (
	"fmt"
	"time"

	"github.com/PagerDuty/go-pagerduty"
)

// RenderedScheduleEntry is a different representation of the
// pagerduty.RenderedScheduleEntry type. The notable changes are that the Start
// and End fields are time.Time objects, rather than RFC3337 timestamp strings,
// and that schedule of the schedule the entry came from is included in the
// Schedule field, for easier grouping.
type RenderedScheduleEntry struct {
	Start    time.Time
	End      time.Time
	User     pagerduty.APIObject
	Schedule string
}

// GetRenderedScheduleEntries returns a list of RenderedScheduleEntry objects
// corresponding to the final rendered schedule. It returns an error if any of
// the Start or End fields fail to parse as RFC3337 timestamps.
func GetRenderedScheduleEntries(schedule *pagerduty.Schedule) ([]*RenderedScheduleEntry, error) {
	if schedule == nil {
		return nil, nil
	}

	entries := schedule.FinalSchedule.RenderedScheduleEntries
	results := make([]*RenderedScheduleEntry, len(entries))

	for i, entry := range entries {
		se, err := ParseRenderedScheduleEntry(schedule, entry)
		if err != nil {
			return nil, err
		}

		results[i] = se
	}

	return results, nil
}

// ParseRenderedScheduleEntry parses a RenderedScheduleEntry object out of a
// pagerduty.RenderedScheduleEntry object. It returns an error if either the
// Start or End fields fail to parse as RFC337 timestamps.
func ParseRenderedScheduleEntry(schedule *pagerduty.Schedule, entry pagerduty.RenderedScheduleEntry) (*RenderedScheduleEntry, error) {
	start, err := time.Parse(time.RFC3339, entry.Start)
	if err != nil {
		return nil, fmt.Errorf("could not parse start time: %w", err)
	}

	end, err := time.Parse(time.RFC3339, entry.End)
	if err != nil {
		return nil, fmt.Errorf("could not parse end time: %w", err)
	}

	var name string
	if schedule != nil {
		name = schedule.Name
	}

	return &RenderedScheduleEntry{
		Start:    start,
		End:      end,
		User:     entry.User,
		Schedule: name,
	}, nil
}

// RenderedScheduleEntries is a list of RenderedScheduleEntry objects.
type RenderedScheduleEntries []*RenderedScheduleEntry

// GroupBy returns a mapping of some string key, as determined by the provided
// keyFn, to lists of RenderedScheduleEntry objects that map to that key.
func (entries RenderedScheduleEntries) GroupBy(keyFn func(entry *RenderedScheduleEntry) string) map[string][]*RenderedScheduleEntry {
	groupedEntries := map[string][]*RenderedScheduleEntry{}

	for _, entry := range entries {
		key := keyFn(entry)

		if _, ok := groupedEntries[key]; !ok {
			groupedEntries[key] = []*RenderedScheduleEntry{entry}
			continue
		}

		groupedEntries[key] = append(groupedEntries[key], entry)
	}

	return groupedEntries
}
