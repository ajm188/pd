package pd

import (
	"fmt"
	"time"

	"github.com/PagerDuty/go-pagerduty"
)

type RenderedScheduleEntry struct {
	Start    time.Time
	End      time.Time
	User     pagerduty.APIObject
	Schedule string
}

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
