# parallelo

Execute the same command in sevaral threads.

## Usage

```BASH
Run the same program in several threads

Usage: ./parallelo [option ...] [program [args] ]
  -concurrency int
        Number of threads (default 50)
  -p int
        Stats message printing interval in seconds (default 5)
  -throttle int
        Throttle between execution in milliseconds (default 1)
  -v    Verbose logging
```

## Installation

```BASH
make build
```
