package svlog

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"svlogj/pkg/types"
	"svlogj/pkg/utils"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
)

type SvLogger struct {
	LineHandler   func(info types.Info)
	ParseConfig   types.ParseConfig
	Fifo          utils.Fifo[types.Info]
	printedLineNr int // non consecutive prints show block separation
	matchLineNr   int // the line nr that grepped true
	bootTime      time.Time
}

func Svlog(parseConfig types.ParseConfig) {
	logger := SvLogger{
		ParseConfig: parseConfig,
		matchLineNr: -1,
	}
	logger.LineHandler = logger.maybePrintLine
	logger.ParseLog()
}

func (l *SvLogger) parseBootTime() {
	file, err := os.Open("/proc/uptime")
	cobra.CheckErr(err)
	defer func(file *os.File) {
		err := file.Close()
		cobra.CheckErr(err)
	}(file)
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	line := scanner.Text()
	uptime, _ := strconv.ParseFloat(strings.Split(line, " ")[0], 64)
	duration := time.Duration(-uptime * float64(time.Second))
	l.bootTime = time.Now().Add(duration)
}

func (l *SvLogger) ParseLog() {

	l.parseBootTime()

	if l.ParseConfig.Grep.Before != 0 {
		l.Fifo = utils.NewFifo[types.Info](l.ParseConfig.Grep.Before)
	}

	linePattern := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+) ((\w+)\.(\w+):)?(.*).*$`)

	var cmd *exec.Cmd
	if len(l.ParseConfig.Service) == 0 {
		cmd = exec.Command("stdbuf", "-oL", "svlogtail")
	} else {
		cmd = exec.Command("stdbuf", "-oL", "svlogtail", l.ParseConfig.Service)
	}
	pipe, _ := cmd.StdoutPipe()
	defer func(pipe io.ReadCloser) {
		_ = pipe.Close()
	}(pipe)
	cobra.CheckErr(cmd.Start())

	scanner := bufio.NewScanner(pipe)
	var running atomic.Bool
	// go routine to Check if the svlogtail has stopped.
	if !l.ParseConfig.Follow {
		go func() {
			for {
				running.Store(false)
				time.Sleep(200 * time.Millisecond)
				// the main loop should have kept the running to true
				if !running.Load() {
					cobra.CheckErr(pipe.Close())
					break
				}
			}
		}()
	}
	lineNr := 0
	for scanner.Scan() {
		line := scanner.Text()
		m := linePattern.FindStringSubmatch(line)
		if m == nil {
			fmt.Printf("CANT HANDLE %s\n", line)
			continue
		}
		t, err := time.Parse(`2006-01-02T15:04:05.99999`, m[1])
		if err != nil {
			fmt.Printf("CANT PARSE TIMESTAMP %s\n", m[1])
			continue
		}
		info := types.Info{
			Timestamp: t,
			Facility:  m[3],
			Level:     m[4],
			Message:   m[5],
			LineNr:    lineNr,
		}
		guessEntityAndPid(&info)
		l.LineHandler(info)
		running.Store(true) // for the 'follow' functionality
		lineNr += 1
	}
}

// guessEntityAndPid Heuristically tries to find the entity name and PID from the message line.
//
// Not perfect but seems to be able to find entities like NetworkManager, dbus, ...
func guessEntityAndPid(info *types.Info) {
	// kernel messages have no pattern whatsoever. Leave the default values
	if info.Facility == "kern" {
		return
	}
	// look for entity-name[pid]
	entityPat := regexp.MustCompile(`([\w-]+)\[(\d+)]`)
	// look for entity-name:
	entityPat2 := regexp.MustCompile(`[, \t]([a-zA-Z][a-zA-Z0-9-]+):`)
	entity := entityPat.FindStringSubmatch(info.Message)
	if entity != nil {
		pid, _ := strconv.Atoi(entity[2])
		info.Entity, info.Pid = entity[1], pid
	} else {
		entity = entityPat2.FindStringSubmatch(info.Message)
		if entity != nil {
			info.Entity = entity[1]
		}
	}
}

// printField print a field value, optionally coloring
func (l *SvLogger) printField(value string, selector string, formatString string) {
	if !l.ParseConfig.Monochrome && value == selector {
		formatString = "\033[" + l.ParseConfig.AnsiColor + "m" + formatString + "\033[0m"
	}
	_, _ = fmt.Printf(formatString, value)
}

// printLine print one log line
func (l *SvLogger) printLine(i types.Info) {
	if l.printedLineNr > 0 && i.LineNr != l.printedLineNr+1 {
		// print separator line if non-consecutive
		_, _ = fmt.Printf("---\n")
	}
	_, _ = fmt.Printf("%s ", l.formatTimestamp(i))
	l.printField(i.Facility, l.ParseConfig.Facility, "%6s.")
	l.printField(i.Level, l.ParseConfig.Level, "%-6s ")
	l.printField(i.Entity, l.ParseConfig.Entity, "%s")
	_, _ = fmt.Printf(" (%d) %s\n", i.Pid, i.Message)
	l.printedLineNr = i.LineNr
}

// maybePrintLine print a log line if the conditions match
//
// will also handle grep BEFORE and AFTER and CONTEXT values
func (l *SvLogger) maybePrintLine(info types.Info) {
	conf := l.ParseConfig
	matched := (len(conf.Entity) == 0 && len(conf.Level) == 0 && len(conf.Facility) == 0) ||
		len(conf.Entity) != 0 && info.Entity == conf.Entity ||
		len(conf.Level) != 0 && info.Level == conf.Level ||
		len(conf.Facility) != 0 && info.Facility == conf.Facility

	if matched {
		l.matchLineNr = info.LineNr
	}
	if matched && l.Fifo.Fill > 0 {
		// handles the grep BEFORE case
		for {
			v, ok := l.Fifo.Get()
			if !ok {
				break
			}
			l.printLine(v)
		}
	}
	if matched {
		// we're on the line of the match. Just print it
		l.printLine(info)
	} else {
		// we might need to handle the grep AFTER case
		if conf.Grep.After > 0 && l.matchLineNr >= 0 && (info.LineNr-l.matchLineNr) <= conf.Grep.After {
			// handles the grep AFTER case
			l.printLine(info)
		} else {
			// when not printing just fill the fifo if there's a grep BEFORE or CONTEXT
			if l.Fifo.Cap > 0 {
				l.Fifo.Push(info)
			}
		}
	}
}

// return a timestamp based on the configuration
func (l *SvLogger) formatTimestamp(info types.Info) string {
	switch l.ParseConfig.TimeConfig {
	case "uptime_s":
		seconds := info.Timestamp.Sub(l.bootTime).Seconds()
		return fmt.Sprintf("%9.03fs", seconds)
	case "local":
		location, _ := time.LoadLocation("Europe/Madrid")
		return info.Timestamp.In(location).Format(time.RFC3339)
	}
	return info.Timestamp.Format(time.RFC3339)
}
