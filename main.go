package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/spf13/cobra"
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
	for _, schedID := range scheduleIDs {
		sched, err := client.GetSchedule(schedID, pagerduty.GetScheduleOptions{
			Since: now.Add(since).Format(time.RFC3339),
			Until: now.Add(until).Format(time.RFC3339),
		})
		if err != nil {
			return err
		}

		data, err := json.Marshal(sched.FinalSchedule.RenderedScheduleEntries)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", data)
	}

	return nil
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
