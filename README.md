<p align="center">
  <h1 align="center">agendexpr</h1>
  <p align="center">A high-performance, zero-allocation Go library for cron-like expression parsing.</p>
</p>

<p align="center">
  <a href="http://godoc.org/github.com/hanzjefferson/agendexpr"><img src="https://img.shields.io/badge/go.dev-reference-blue?logo=go" alt="Go Reference"></a>
  <a href="https://github.com/hanzjefferson/agendexpr/blob/main/LICENSE"><img src="https://img.shields.io/github/license/hanzjefferson/agendexpr" alt="License"></a>
  <img src="https://img.shields.io/badge/go-1.26-blue" alt="Go 1.26">
</p>

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Expression Format](#expression-format)
- [API Reference](#api-reference)
- [Named Aliases](#named-aliases)
- [Benchmarks](#benchmarks)
- [Credits](#credits)
- [Contributing](#contributing)
- [License](#license)

---

## Features

- **6-field expression** — seconds, minutes, hours, day-of-month, month, day-of-week
- **Bitmask storage** — O(1) match checks on every field
- **`Next()` scheduling** — find the next matching timestamp from any `time.Time`
- **`Previous()` lookback** — find the latest matching timestamp before any `time.Time`
- **`Due()` checking** — test if a timestamp matches the expression (second to month)
- **`Countdown()`** — seconds until the next match
- **Name normalization** — `JAN`–`DEC`, `SUN`–`SAT` in any case
- **Named aliases** — `@hourly`, `@weekly`, `@monthly`, `@ticks`, etc.
- **Zero heap allocs** in parsing and matching hot paths
- **No external dependencies** beyond stdlib

## Installation

```bash
go get github.com/hanzjefferson/agendexpr
```

## Quick Start

```go
package main

import (
	"fmt"
	"time"

	"github.com/hanzjefferson/agendexpr"
)

func main() {
	// Every day at 08:30:00
	expr := agendexpr.MustParse("0 30 8 * * *")

	now := time.Now()
	next := expr.Next(now)
	fmt.Println("Next run:", next)

	due := expr.Due(next)
	fmt.Println("Is due:", due)

	countdown := expr.Countdown(now)
	fmt.Println("Seconds until next:", countdown)
}
```

## Expression Format

Six space-separated fields:

```
┌───────── second (0–59)
│ ┌───────── minute (0–59)
│ │ ┌───────── hour   (0–23)
│ │ │ ┌───────── day-of-month (1–31)
│ │ │ │ ┌───────── month (1–12, JAN–DEC)
│ │ │ │ │ ┌───────── day-of-week (0–6, SUN–SAT)
│ │ │ │ │ │
* * * * * *
```

Each field supports:

| Notation     | Example       | Description                    |
|--------------|---------------|--------------------------------|
| `*`          | `*`           | Every value                    |
| `N`          | `5`           | Single value                   |
| `N,M,...`    | `1,15,30`     | Multiple values                |
| `N-M`        | `1–5`         | Inclusive range                |
| `*/S`        | `*/15`        | Step over full range           |
| `N-M/S`      | `10–20/5`     | Step over a range              |

Month and day-of-week names are case-insensitive: `jan`, `JAN`, `Jan` all resolve to `1`.

## API Reference

Full API reference is available [here](http://godoc.org/github.com/hanzjefferson/agendexpr)

## Named Aliases

| Alias        | Expression       | Description       |
| ------------ | ---------------- | ----------------- |
| `@ticks`     | `* * * * * *`    | Every second      |
| `@minutely`  | `0 * * * * *`    | Every minute      |
| `@5minutes`  | `0 */5 * * * *`  | Every 5 minutes   |
| `@10minutes` | `0 */10 * * * *` | Every 10 minutes  |
| `@30minutes` | `0 */30 * * * *` | Every 30 minutes  |
| `@hourly`    | `0 0 * * * *`    | Every hour        |
| `@daily`     | `0 0 0 * * *`    | Every day         |
| `@weekly`    | `0 0 0 * * 0`    | Every week        |
| `@monthly`   | `0 0 0 1 * *`    | Every month       |

## Benchmarks

All benchmarks are run with **randomized expressions** (1000 unique patterns) to reflect real-world usage.

### Performance Results

| Benchmark Case       | Execution Time | Memory Allocation | Allocations Count |
| -------------------- | -------------- | ----------------- | ----------------- |
| `Parse`              | `~550 ns/op`   | `0 B/op`          | `0 allocs/op`     |
| `Parse` (randomized) | `~1060 ns/op`  | `0 B/op`          | `0 allocs/op`     |
| `Next`               | `~1440 ns/op`  | `0 B/op`          | `0 allocs/op`     |
| `Next` (randomized)  | `~10400 ns/op` | `0 B/op`          | `0 allocs/op`     |
| `Due`                | `~118 ns/op`   | `0 B/op`          | `0 allocs/op`     |
| `Due` (randomized)   | `~84 ns/op`    | `0 B/op`          | `0 allocs/op`     |

**NOTES:** _All benchmarks achieved **zero allocations** (`0 allocs/op`), ensuring minimal GC (Garbage Collection) pressure during execution. Randomized benchmarks use 1000 unique cron-like expressions with varied notations (wildcards, ranges, steps, lists) to simulate production workloads._

## Credits

- **Cron expression format** — based on the standard cron syntax as documented on [Wikipedia](https://en.wikipedia.org/wiki/Cron#CRON_expression)
- **[gorhill/cronexpr](https://github.com/gorhill/cronexpr)** — inspired by this cron expression parser for Go, but i demonstrated the bitmask-based approach for efficient time matching

## Contributing

Contributions are welcome. Open an issue or pull request on [GitHub](https://github.com/hanzjefferson/agendexpr).

## License

MIT &mdash; see [LICENSE](https://github.com/hanzjefferson/agendexpr/blob/main/LICENSE).
