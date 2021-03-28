package pd

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/PagerDuty/go-pagerduty"
)

// GetAllSchedules makes concurrent client.GetSchedule calls for each schedule
// ID. Unlike GetSchedules, if any GetSchedule call fails, this method returns
// an error. Consequently, callers should use this function when they do not
// want to permit partial responses.
func GetAllSchedules(client *pagerduty.Client, scheduleIDs []string, opts pagerduty.GetScheduleOptions) ([]*pagerduty.Schedule, error) {
	results, failures, err := GetSchedules(client, scheduleIDs, opts)
	if err != nil {
		return nil, err
	}

	if len(failures) > 0 {
		return nil, errors.New(strings.Join(failures, "; "))
	}

	return results, nil
}

// GetSchedules makes concurrent client.GetSchedule calls for each schedule ID.
// Individual call failures are collected into a failures slice, so that callers
// can decide whether to permit partial responses or not. If every GetSchedule
// call returns an error, this method will return an error, which is the
// concatenation of those failure messages.
func GetSchedules(client *pagerduty.Client, scheduleIDs []string, opts pagerduty.GetScheduleOptions) ([]*pagerduty.Schedule, []string, error) {
	results := make([]*pagerduty.Schedule, 0, len(scheduleIDs))
	failures := make([]string, 0, len(scheduleIDs))

	var (
		m  sync.Mutex
		wg sync.WaitGroup
	)

	for _, scheduleID := range scheduleIDs {
		wg.Add(1)

		go func(id string) {
			defer wg.Done()

			sched, err := client.GetSchedule(id, opts)

			m.Lock()
			defer m.Unlock()

			if err != nil {
				failures = append(failures, fmt.Sprintf("%s: %s", id, err))
				return
			}

			results = append(results, sched)
		}(scheduleID)
	}

	wg.Wait()

	switch len(failures) {
	case 0:
		return results, nil, nil
	case len(scheduleIDs):
		return nil, failures, errors.New(strings.Join(failures, "; "))
	default:
		return results, nil, nil
	}
}
