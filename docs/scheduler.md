# Scheduler

The `laravel-go` library provides a robust task scheduler inspired by Laravel's schedule. It supports:
- Cron expression scheduling
- Overlapping prevention
- Distributed locks (`OnOneServer`) using Redis or Database

## Commands

To run the scheduler, use the `schedule:run` command. This command will start the scheduler and block until interrupted (SIGINT/SIGTERM).

```bash
go run main.go schedule:run
```

## Configuration

The scheduler automatically uses the `CACHE_STORE` configuration from your `.env` file to determine the lock provider for `OnOneServer` tasks.

- `CACHE_STORE=redis`: Uses Redis locks (requires `REDIS_*` config).
- `CACHE_STORE=database`: Uses Database locks (requires `DB_CONNECTION` config).
- Default: No lock provider (local only).

## Registering Tasks

You should register your scheduled tasks in your application's entry point (e.g., `main.go`) before executing the root command.

```go
package main

import (
    "fmt"
    "github.com/pixelvide/laravel-go/pkg/schedule"
    "github.com/pixelvide/laravel-go/pkg/root"

    // Import console to register commands
    _ "github.com/pixelvide/laravel-go/pkg/console"
)

func main() {
    // 1. Simple Cron Job
    // Runs every minute
    schedule.Register("* * * * *", func() {
        fmt.Println("This runs every minute")
    })

    // 2. Prevent Overlapping (Local)
    // If the task takes longer than 1 minute, the next run is skipped.
    schedule.Register("* * * * *", func() {
        // Heavy task...
    }, schedule.WithoutOverlapping())

    // 3. Distributed Lock (OnOneServer)
    // Ensures the task runs on only ONE server in your cluster.
    // Requires CACHE_STORE to be configured (redis or database).
    schedule.Register("0 0 * * *", func() {
        fmt.Println("Daily Cleanup")
    }, schedule.OnOneServer("daily-cleanup"))

    root.Execute()
}
```

## Database Locking

If you choose `CACHE_STORE=database`, the scheduler uses:
- **MySQL**: `GET_LOCK(name, 0)` / `RELEASE_LOCK(name)`
- **PostgreSQL**: `pg_try_advisory_lock(key)` / `pg_advisory_unlock(key)`

Ensure your database user has permission to use these locking functions.
