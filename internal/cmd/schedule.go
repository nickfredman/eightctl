package cmd

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/steipete/eightctl/internal/client"
	"github.com/steipete/eightctl/internal/output"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Manage device temperature schedules (cloud)",
}

var scheduleNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Show next upcoming schedule events",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuthFields(); err != nil {
			return err
		}
		tzName := viper.GetString("timezone")
		loc := time.Local
		if tzName != "" && tzName != "local" {
			l, err := time.LoadLocation(tzName)
			if err != nil {
				return err
			}
			loc = l
		}
		now := time.Now().In(loc)

		cl := client.New(viper.GetString("email"), viper.GetString("password"), viper.GetString("user_id"), viper.GetString("client_id"), viper.GetString("client_secret"))
		scheds, err := cl.ListSchedules(context.Background())
		if err != nil {
			return err
		}

		rows := make([]map[string]any, 0, len(scheds))
		for _, s := range scheds {
			next := nextOccurrence(now, s, loc)
			rows = append(rows, map[string]any{
				"id":      s.ID,
				"start":   s.StartTime,
				"days":    s.DaysOfWeek,
				"level":   s.Level,
				"enabled": s.Enabled,
				"next":    next.Format(time.RFC3339),
			})
		}

		sort.Slice(rows, func(i, j int) bool { return rows[i]["next"].(string) < rows[j]["next"].(string) })
		rows = output.FilterFields(rows, viper.GetStringSlice("fields"))
		headers := []string{"id", "start", "days", "level", "enabled", "next"}
		if len(viper.GetStringSlice("fields")) > 0 {
			headers = viper.GetStringSlice("fields")
		}
		return output.Print(output.Format(viper.GetString("output")), headers, rows)
	},
}

var scheduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List schedules",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuthFields(); err != nil {
			return err
		}
		cl := client.New(viper.GetString("email"), viper.GetString("password"), viper.GetString("user_id"), viper.GetString("client_id"), viper.GetString("client_secret"))
		scheds, err := cl.ListSchedules(context.Background())
		if err != nil {
			return err
		}
		rows := make([]map[string]any, 0, len(scheds))
		for _, s := range scheds {
			rows = append(rows, map[string]any{
				"id":      s.ID,
				"start":   s.StartTime,
				"level":   s.Level,
				"days":    s.DaysOfWeek,
				"enabled": s.Enabled,
			})
		}
		rows = output.FilterFields(rows, viper.GetStringSlice("fields"))
		return output.Print(output.Format(viper.GetString("output")), []string{"id", "start", "level", "days", "enabled"}, rows)
	},
}

var scheduleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuthFields(); err != nil {
			return err
		}
		start := viper.GetString("start")
		if start == "" {
			return fmt.Errorf("--start HH:MM required")
		}
		level := viper.GetInt("level")
		days := viper.GetIntSlice("days")
		if len(days) == 0 {
			return fmt.Errorf("--days required")
		}
		enabled := !viper.GetBool("disabled")
		cl := client.New(viper.GetString("email"), viper.GetString("password"), viper.GetString("user_id"), viper.GetString("client_id"), viper.GetString("client_secret"))
		s := client.TemperatureSchedule{StartTime: start, Level: level, DaysOfWeek: days, Enabled: enabled}
		res, err := cl.CreateSchedule(context.Background(), s)
		if err != nil {
			return err
		}
		fmt.Printf("created schedule %s\n", res.ID)
		return nil
	},
}

var scheduleUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update schedule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuthFields(); err != nil {
			return err
		}
		patch := map[string]any{}
		if cmd.Flags().Changed("start") {
			patch["startTime"] = viper.GetString("start")
		}
		if cmd.Flags().Changed("level") {
			patch["level"] = viper.GetInt("level")
		}
		if cmd.Flags().Changed("days") {
			patch["daysOfWeek"] = viper.GetIntSlice("days")
		}
		if cmd.Flags().Changed("enabled") {
			patch["enabled"] = viper.GetBool("enabled")
		}
		if len(patch) == 0 {
			return fmt.Errorf("no fields to update")
		}
		cl := client.New(viper.GetString("email"), viper.GetString("password"), viper.GetString("user_id"), viper.GetString("client_id"), viper.GetString("client_secret"))
		if _, err := cl.UpdateSchedule(context.Background(), args[0], patch); err != nil {
			return err
		}
		fmt.Println("updated")
		return nil
	},
}

var scheduleDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete schedule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuthFields(); err != nil {
			return err
		}
		cl := client.New(viper.GetString("email"), viper.GetString("password"), viper.GetString("user_id"), viper.GetString("client_id"), viper.GetString("client_secret"))
		if err := cl.DeleteSchedule(context.Background(), args[0]); err != nil {
			return err
		}
		fmt.Println("deleted")
		return nil
	},
}

func init() {
	scheduleCreateCmd.Flags().String("start", "", "HH:MM start time")
	scheduleCreateCmd.Flags().Int("level", 0, "Temperature level -100..100")
	scheduleCreateCmd.Flags().IntSlice("days", nil, "Comma-separated days 0=Sun..6=Sat")
	scheduleCreateCmd.Flags().Bool("disabled", false, "Create disabled")
	viper.BindPFlag("start", scheduleCreateCmd.Flags().Lookup("start"))
	viper.BindPFlag("level", scheduleCreateCmd.Flags().Lookup("level"))
	viper.BindPFlag("days", scheduleCreateCmd.Flags().Lookup("days"))
	viper.BindPFlag("disabled", scheduleCreateCmd.Flags().Lookup("disabled"))

	scheduleUpdateCmd.Flags().String("start", "", "HH:MM start time")
	scheduleUpdateCmd.Flags().Int("level", 0, "Temperature level -100..100")
	scheduleUpdateCmd.Flags().IntSlice("days", nil, "Comma-separated days 0=Sun..6=Sat")
	scheduleUpdateCmd.Flags().Bool("enabled", true, "Enable/disable schedule")
	viper.BindPFlag("start", scheduleUpdateCmd.Flags().Lookup("start"))
	viper.BindPFlag("level", scheduleUpdateCmd.Flags().Lookup("level"))
	viper.BindPFlag("days", scheduleUpdateCmd.Flags().Lookup("days"))
	viper.BindPFlag("enabled", scheduleUpdateCmd.Flags().Lookup("enabled"))

	scheduleCmd.AddCommand(scheduleListCmd, scheduleCreateCmd, scheduleUpdateCmd, scheduleDeleteCmd, scheduleNextCmd)
}

func nextOccurrence(now time.Time, s client.TemperatureSchedule, loc *time.Location) time.Time {
	hour, min, _ := time.Now().Clock()
	if t, err := time.Parse("15:04", s.StartTime); err == nil {
		hour, min, _ = t.Clock()
	}
	days := map[int]bool{}
	for _, d := range s.DaysOfWeek {
		days[d] = true
	}
	for i := 0; i < 14; i++ {
		day := now.In(loc).AddDate(0, 0, i)
		if len(days) > 0 && !days[int(day.Weekday())] {
			continue
		}
		cand := time.Date(day.Year(), day.Month(), day.Day(), hour, min, 0, 0, loc)
		if cand.After(now) {
			return cand
		}
	}
	return now
}
