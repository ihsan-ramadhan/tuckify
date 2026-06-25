package scheduler

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ihsan-ramadhan/tuckify/internal/config"
	"github.com/ihsan-ramadhan/tuckify/internal/organizer"
	"github.com/robfig/cron/v3"
)

func Run(folder, expr string, cfg *config.Config) error {
	c := cron.New()

	_, err := c.AddFunc(expr, func() {
		fmt.Printf("[%s] running organizer on %s\n", time.Now().Format("2006-01-02 15:04:05"), folder)
		results, err := organizer.Organize(folder, cfg, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}
		moved := 0
		for _, r := range results {
			if r.Skipped {
				fmt.Fprintf(os.Stderr, "skipped %s: %s\n", r.Source, r.SkipReason)
				continue
			}
			fmt.Printf("moved %q → %s\n", r.Source, r.Destination)
			moved++
		}
		fmt.Printf("%d file(s) moved\n", moved)
	})
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	c.Start()
	fmt.Printf("scheduler started — press Ctrl+C to stop\n")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	c.Stop()
	fmt.Println("scheduler stopped")
	return nil
}
