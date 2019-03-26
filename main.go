package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Runtime options
var (
	concurrency          int
	verbose              bool
	statsIntervalSeconds int
	// throttle between execCommandution
	throttleMilliseconds int
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})

	flag.IntVar(&concurrency, "concurrency", 50,
		"Number of threads")
	flag.IntVar(&statsIntervalSeconds, "p", 5,
		"Stats message printing interval in seconds")
	flag.IntVar(&throttleMilliseconds, "throttle", 1,
		"Throttle between execution in milliseconds")
	flag.BoolVar(&verbose, "v", false,
		"Verbose logging")
	flag.Usage = func() {
		fmt.Printf(strings.Join([]string{
			"Run the same program in several threads",
			"",
			"Usage: ./parallelo [option ...] [program [args] ]",
			"",
		}, "\n"))
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if verbose {
		log.SetLevel(log.DebugLevel)
	}
	// We need at least one target domain
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	// all remaining parameters are treated as program and args
	execCommand := make([]string, flag.NArg())
	for index, element := range flag.Args() {
		execCommand[index] = element
	}
	log.WithFields(log.Fields{
		"execCommand": strings.Join(execCommand, " "),
	}).Info("Started")
	ch := make(chan []string)

	// Run concurrently
	for threadID := 0; threadID < concurrency; threadID++ {
		go func() {
			for command := range ch {
				cmd := exec.Command(command[0], command[1:]...)
				cmdStdoutReader, _ := cmd.StdoutPipe()
				cmdStderrReader, _ := cmd.StderrPipe()
				go func() {
					scanner := bufio.NewScanner(cmdStdoutReader)
					for scanner.Scan() {
						log.WithField("stdout", scanner.Text()).Debug("got stdout")
					}
				}()
				go func() {
					scanner := bufio.NewScanner(cmdStderrReader)
					for scanner.Scan() {
						log.WithField("stderr", scanner.Text()).Error("got stderr")
					}
				}()
				cmd.Start()
				if err := cmd.Wait(); err != nil {
					log.Errorf("failed to exec command: %v", err)
				}
			}
		}()
	}

	// print statistic in background
	execCnt := 0
	lastPrintedCnt := 0
	go func() {
		for {
			log.Infof("Done %d executions, ~%.2f exec/sec", execCnt, float64((execCnt-lastPrintedCnt)/statsIntervalSeconds))
			lastPrintedCnt = execCnt
			time.Sleep(time.Second * time.Duration(statsIntervalSeconds))
		}
	}()

	// start sendind requests to channel
	throttle := time.Tick(time.Millisecond * time.Duration(throttleMilliseconds))
	for {
		<-throttle
		ch <- execCommand
		execCnt++

	}
}
