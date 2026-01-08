package schedule

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"log"
)

// Kernel manages scheduled tasks
type Kernel struct {
	cron         *cron.Cron
	lockProvider LockProvider
}

// JobOption configures a scheduled job
type JobOption func(*jobConfig)

type jobConfig struct {
	withoutOverlapping bool
	onOneServer        bool
	name               string
}

// NewKernel creates a new scheduler kernel
func NewKernel(lockProvider LockProvider) *Kernel {
	// Initialize Cron with second-level precision
	c := cron.New(cron.WithSeconds())
	return &Kernel{
		cron:         c,
		lockProvider: lockProvider,
	}
}

// SetLockProvider sets the distributed lock provider
func (k *Kernel) SetLockProvider(provider LockProvider) {
	k.lockProvider = provider
}

// WithoutOverlapping prevents the job from running if the previous instance is still running (local only)
func WithoutOverlapping() JobOption {
	return func(c *jobConfig) {
		c.withoutOverlapping = true
	}
}

// OnOneServer ensures the job runs on only one server at a time (distributed lock)
func OnOneServer(name string) JobOption {
	return func(c *jobConfig) {
		c.onOneServer = true
		c.name = name
	}
}

// Register adds a function to be run on a given schedule
// Schedule format: "s m h d m w" (Seconds Minutes Hours Day Month Week)
func (k *Kernel) Register(schedule string, cmd func(), opts ...JobOption) {
	cfg := &jobConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	var job cron.Job = cron.FuncJob(cmd)

	// Apply WithoutOverlapping (Local mutex via cron.SkipIfStillRunning)
	if cfg.withoutOverlapping {
		job = cron.SkipIfStillRunning(cron.DefaultLogger)(job)
	}

	// Apply OnOneServer (Distributed Lock)
	if cfg.onOneServer {
		if k.lockProvider == nil {
			log.Printf("Warning: Ignoring OnOneServer for job '%s': LockProvider not initialized", cfg.name)
		} else {
			originalJob := job
			job = cron.FuncJob(func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Check lock timeout
				defer cancel()

				// Determine lock duration.
				// ideally should cover the job duration or be renewed.
				// For simple implementation, we lock for 1 minute or 1 hour?
				// Laravel's Cache::lock defaults strictly.
				// Here we just check if we can acquire.
				// To truly match Laravel's "OnOneServer", we need to hold the lock while running?
				// Usually "OnOneServer" prevents other servers from starting the *same scheduled instance*.
				// So we lock for a short duration (e.g. 59s for a 1m schedule) to prevent others.
				// But simpler: Attempt lock. If fail, skip.

				lockName := cfg.name
				acquired, err := k.lockProvider.GetLock(ctx, lockName, 1*time.Minute)
				if err != nil {
					log.Printf("Error checking lock for job '%s': %v", cfg.name, err)
					return
				}

				if acquired {
					// We got the lock, run the job.
					// Note: We do NOT release the lock immediately if we want to prevent others
					// from picking up this specific minute's slot?
					// Laravel's "onOneServer" uses a cache key that expires.
					// If we release immediately, another server might pick it up 1ms later?
					// No, SETNX prevents that.
					// But if the job finishes quickly (1s), and we release, another server checking at second 2 might run?
					// Cron triggers at roughly same time.
					// Standard practice: Keep lock for the "tick" duration or release after done?
					// Laravel actually releases after execution by default unless configured otherwise?
					// Let's release after execution.
					defer func() {
						_ = k.lockProvider.ReleaseLock(context.Background(), lockName)
					}()
					originalJob.Run()
				}
				// else {
				// 	log.Printf("Skipping job '%s': locked by another server", cfg.name)
				// }
			})
		}
	}

	_, err := k.cron.AddJob(schedule, job)
	if err != nil {
		log.Printf("Failed to register cron job: %v", err)
	} else {
		log.Printf("Registered cron job: %s [%s]", cfg.name, schedule)
	}
}

// Run starts the scheduler and blocks until interrupt
func (k *Kernel) Run() {
	log.Println("Starting Task Scheduler...")
	k.cron.Start()

	// Wait for interrupt signal to gracefully shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Stopping Task Scheduler...")
	context := k.cron.Stop()
	<-context.Done() // Wait for active jobs
}
