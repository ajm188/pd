package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/spf13/cobra"

	"github.com/ajm188/pd"
)

var (
	authToken   string
	scheduleIDs []string
	since       time.Duration
	until       time.Duration

	rootCmd = &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			if until <= since {
				return fmt.Errorf("--until (%v) must be larger than --since (%v)", until, since)
			}

			cmd.SilenceErrors = true
			return run()
		},
	}
)

func run() error {
	now := time.Now().UTC()

	client := pagerduty.NewClient(authToken)

	entriesByUser := map[string][]*pd.RenderedScheduleEntry{}

	for _, schedID := range scheduleIDs {
		sched, err := client.GetSchedule(schedID, pagerduty.GetScheduleOptions{
			Since: now.Add(since).Format(time.RFC3339),
			Until: now.Add(until).Format(time.RFC3339),
		})
		if err != nil {
			// TODO: log and skip (surface back to caller ... somehow)
			return err
		}

		entries, err := pd.GetRenderedScheduleEntries(sched)
		if err != nil {
			// TODO: log and skip (surface back to caller ... somehow)
			return err
		}

		for _, entry := range entries {
			userEntries, ok := entriesByUser[entry.User.ID]
			if !ok {
				userEntries = []*pd.RenderedScheduleEntry{}
			}

			entriesByUser[entry.User.ID] = append(userEntries, entry)
		}

		data, err := json.Marshal(entries)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", data)
	}

	findConflicts(entriesByUser)

	return nil
}

func findConflicts(entriesByUser map[string][]*pd.RenderedScheduleEntry) {
	var wg sync.WaitGroup

	for userID, entries := range entriesByUser {
		wg.Add(1)

		go func(userID string, entries []*pd.RenderedScheduleEntry) {
			defer wg.Done()

			conflicts := [][2]*pd.RenderedScheduleEntry{}

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

					log.Printf("CONFLICT: %s is in both %s and %s from %s to %s\n", left.User.Summary, left.Schedule, right.Schedule, right.Start, left.End)

					conflicts = append(conflicts, [2]*pd.RenderedScheduleEntry{left, right})
				}
			}
		}(userID, entries)
	}

	wg.Wait()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	// flags:
	// - start time to search
	// - end time to search
	// - list of schedules to cross-reference
	rootCmd.Flags().StringVar(&authToken, "auth-token", "", "auth token (TODO: allow reading from file instead)")
	rootCmd.Flags().StringSliceVarP(&scheduleIDs, "schedule", "s", nil, "schedule IDs to check")
	rootCmd.Flags().DurationVar(&since, "since", 0, "duration offset (relative to time.Now) of the schedules to check; e.g. to go into the past specify '-1h'")
	rootCmd.Flags().DurationVar(&until, "until", time.Hour*24*14, "duration offset (relative to time.Now) of the schedule to check. Must be larger than --since")

	rootCmd.MarkFlagRequired("auth-token")
}
