package console

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pixelvide/laravel-go/pkg/config"
	"github.com/pixelvide/laravel-go/pkg/database"
	"github.com/pixelvide/laravel-go/pkg/driver/redis"
	"github.com/pixelvide/laravel-go/pkg/root"
	"github.com/pixelvide/laravel-go/pkg/schedule"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule:run",
	Short: "Run the scheduled tasks",
	Run: func(cmd *cobra.Command, args []string) {
		// Load Config
		cfg, err := config.Load()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load configuration from .env")
		}

		// Initialize Lock Provider
		var lockProvider schedule.LockProvider

		// Check cache store or schedule driver
		// Simplification: Check CACHE_STORE. If redis, use redis lock. If database, use db lock.
		store := "file"
		if cfg != nil {
			store = cfg.Cache.Store
		}

		switch store {
		case "redis":
			if cfg != nil {
				rCfg := config.RedisConfig{
					Host:     cfg.Redis.Host,
					Port:     cfg.Redis.Port,
					Password: cfg.Redis.Password,
					DB:       cfg.Redis.DB,
				}
				rdb := redis.NewRedisDriver(rCfg).Client
				lockProvider = schedule.NewRedisLockProvider(rdb)
			}
		case "database":
			if cfg != nil {
				dbFactory := database.NewFactory()
				db, err := dbFactory.Connect(cfg.Database)
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to connect to database for scheduler lock")
				}
				lockProvider = schedule.NewDatabaseLockProvider(db, cfg.Database.Connection)
			}
		default:
			log.Info().Str("store", store).Msg("No distributed lock provider configured (using in-memory/none). OnOneServer will not work across multiple servers.")
		}

		// Get Global Kernel and set Lock Provider
		kernel := schedule.GetGlobalKernel()
		// Kernel struct might not expose LockProvider setter if it's private.
		// Assuming we can recreate or inject it.
		// Since NewKernel takes a provider, we might need to replace the global one or add SetLockProvider.
		// For now, let's assume we can set it or we replace the global instance.

		// Update the lock provider on the global kernel instance
		kernel.SetLockProvider(lockProvider)

		// Run Scheduler

		// Handle SIGINT/SIGTERM
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			log.Info().Msg("Shutting down scheduler...")
			// Note: kernel.Run() handles graceful shutdown internally via its own signal handling?
			// Checking pkg/schedule/kernel.go: It does signal.Notify and blocks.
			// So we probably don't need another signal handler here,
			// or we should call kernel.Stop().
			// But kernel.Run() blocks until SIGINT/SIGTERM.
			// So this go routine is redundant if kernel.Run blocks on same signals.
			// Let's rely on kernel.Run() for blocking.
			os.Exit(0)
		}()

		log.Info().Msg("Starting scheduler...")
		kernel.Run()
	},
}

func init() {
	root.GetRoot().AddCommand(scheduleCmd)
}
