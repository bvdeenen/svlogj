package svlog

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"svlogj/pkg/types"
	"svlogj/pkg/utils"
	"sync/atomic"
	"time"
)

type SvLogger struct {
	LineHandler   func(info types.Info)
	ParseConfig   types.ParseConfig
	Follow        bool
	Fifo          utils.Fifo[types.Info]
	printedLineNr int // for block separation
	matchLineNr   int // the line nr that grepped true
	bootTime      time.Time
}

func Svlog(c types.ParseConfig) {
	logger := SvLogger{
		ParseConfig: c,
		Follow:      c.Follow,
		matchLineNr: -1,
	}
	logger.LineHandler = logger.MaybePrintLine
	logger.ParseLog()
}

func (self *SvLogger) parseBootTime() {
	file, err := os.Open("/proc/uptime")
	utils.Check(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	line := scanner.Text()
	uptime, _ := strconv.ParseFloat(strings.Split(line, " ")[0], 64)
	duration := time.Duration(-uptime * float64(time.Second))
	self.bootTime = time.Now().Add(duration)
}

func (self *SvLogger) ParseLog() {

	self.parseBootTime()

	if self.ParseConfig.Grep.Before != 0 {
		self.Fifo = utils.NewFifo[types.Info](self.ParseConfig.Grep.Before)
	}

	line_pattern := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+) ((\w+)\.(\w+):)?(.*).*$`)

	var cmd *exec.Cmd
	if len(self.ParseConfig.Service) == 0 {
		cmd = exec.Command("svlogtail")
	} else {
		cmd = exec.Command("svlogtail", self.ParseConfig.Service)
	}
	pipe, _ := cmd.StdoutPipe()
	defer pipe.Close()
	cmd.Start()
	scanner := bufio.NewScanner(pipe)
	var running atomic.Bool
	// go routine to Check if the svlogtail has stopped.
	if !self.Follow {
		go func() {
			for {
				running.Store(false)
				time.Sleep(200 * time.Millisecond)
				// the main loop should have kept the running to true
				if !running.Load() {
					pipe.Close()
					break
				}
			}
		}()
	}
	lineNr := 0
	for scanner.Scan() {
		line := scanner.Text()
		m := line_pattern.FindStringSubmatch(line)
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
		self.LineHandler(info)
		running.Store(true) // for the 'follow' functionality
		lineNr += 1
	}
}

// Heuristically tries to find the entity name and PID from the message line.
//
// Not perfect but seems to be able to find entities like NetworkManager, dbus, ...
func guessEntityAndPid(info *types.Info) {
	// kernel messages have no pattern whatsoever. Leave the default values
	if info.Facility == "kern" {
		return
	}
	// look for entity-name[pid]
	entity_pat := regexp.MustCompile(`([\w-]+)\[(\d+)\]`)
	// look for entity-name:
	entity_pat2 := regexp.MustCompile(`[, \t]([a-zA-Z][a-zA-Z0-9-]+):`)
	entity := entity_pat.FindStringSubmatch(info.Message)
	if entity != nil {
		pid, _ := strconv.Atoi(entity[2])
		info.Entity, info.Pid = entity[1], pid
	} else {
		entity = entity_pat2.FindStringSubmatch(info.Message)
		if entity != nil {
			info.Entity = entity[1]
		}
	}
}

// print a field value, optionally coloring
func (self *SvLogger) printField(value string, selector string, formatString string) {
	if !self.ParseConfig.Monochrome && value == selector {
		formatString = "\033[" + self.ParseConfig.AnsiColor + "m" + formatString + "\033[0m"
	}
	_, _ = fmt.Printf(formatString, value)
}

// print one log line
func (self *SvLogger) printLine(i types.Info) {
	if self.printedLineNr > 0 && i.LineNr != self.printedLineNr+1 {
		// print separator line if non-consecutive
		_, _ = fmt.Printf("---\n")
	}
	_, _ = fmt.Printf("%s ", self.formatTimestamp(i))
	self.printField(i.Facility, self.ParseConfig.Facility, "%6s.")
	self.printField(i.Level, self.ParseConfig.Level, "%-6s ")
	self.printField(i.Entity, self.ParseConfig.Entity, "%s")
	_, _ = fmt.Printf(" (%d) %s\n", i.Pid, i.Message)
	self.printedLineNr = i.LineNr
}

// print a log line if the conditions match
//
// will also handle grep BEFORE and AFTER and CONTEXT values
func (self *SvLogger) MaybePrintLine(info types.Info) {
	conf := self.ParseConfig
	matched := (len(conf.Entity) == 0 && len(conf.Level) == 0 && len(conf.Facility) == 0) ||
		len(conf.Entity) != 0 && info.Entity == conf.Entity ||
		len(conf.Level) != 0 && info.Level == conf.Level ||
		len(conf.Facility) != 0 && info.Facility == conf.Facility

	if matched {
		self.matchLineNr = info.LineNr
	}
	if matched && self.Fifo.Fill > 0 {
		// handles the grep BEFORE case
		for {
			v, ok := self.Fifo.Get()
			if !ok {
				break
			}
			self.printLine(v)
		}
	}
	if matched {
		// we're on the line of the match. Just print it
		self.printLine(info)
	} else {
		// we might need to handle the grep AFTER case
		if conf.Grep.After > 0 && self.matchLineNr >= 0 && (info.LineNr-self.matchLineNr) <= conf.Grep.After {
			// handles the grep AFTER case
			self.printLine(info)
		} else {
			// when not printing just fill the fifo if there's a grep BEFORE or CONTEXT
			if self.Fifo.Cap > 0 {
				self.Fifo.Push(info)
			}
		}
	}
}

// return a timestamp based on the configuration
func (self *SvLogger) formatTimestamp(info types.Info) string {
	switch self.ParseConfig.TimeConfig {
	case "uptime_s":
		seconds := info.Timestamp.Sub(self.bootTime).Seconds()
		return fmt.Sprintf("%9.03fs", seconds)
	case "local":
		location, _ := time.LoadLocation("Europe/Madrid")
		return info.Timestamp.In(location).Format(time.RFC3339)
	}
	return info.Timestamp.Format(time.RFC3339)
}
