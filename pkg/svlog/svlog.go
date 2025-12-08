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
	selectLineNr  int // the line nr that grepped true
	bootTime      time.Time
}

func Svlog(c types.ParseConfig) {
	logger := SvLogger{
		ParseConfig:  c,
		Follow:       true,
		selectLineNr: -1,
	}
	logger.LineHandler = logger.HandleInterpretedLine
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
	fmt.Printf("%v\n", self.bootTime)

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
				time.Sleep(100 * time.Millisecond)
				// the main loop should have kept the running to true
				if !running.Load() {
					//fmt.Printf("Closing svlogtail because of timeout\n")
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
			Entity:    "",
			Pid:       0,
			Message:   m[5],
			LineNr:    lineNr,
		}
		guessEntityAndPid(&info)
		self.LineHandler(info)
		running.Store(true)
		lineNr += 1
	}
}

func guessEntityAndPid(info *types.Info) {
	// kernel messages have no pattern whatsoever
	if info.Facility == "kern" {
		return
	}
	// Heuristically find the entity that created the message
	entity_pat := regexp.MustCompile(`([\w-]+)\[(\d+)\]`)
	entity_pat2 := regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9-]+):`)
	all_numbers := regexp.MustCompile(`^[0-9-.]+$`)
	entity := entity_pat.FindStringSubmatch(info.Message)
	if entity != nil {
		pid, _ := strconv.Atoi(entity[2])
		info.Entity, info.Pid = entity[1], pid
	} else {
		entity = entity_pat2.FindStringSubmatch(info.Message)
		if entity != nil {
			if all_numbers.FindStringSubmatch(entity[1]) == nil {
				info.Entity = entity[1]
			}
		}
	}
}

func (self *SvLogger) HandleInterpretedLine(info types.Info) {

	parse_config := self.ParseConfig
	printer := func(i types.Info) {
		if self.printedLineNr > 0 && i.LineNr != self.printedLineNr+1 {
			_, _ = fmt.Printf("---\n")
		}
		timestamp := self.displayTimestamp(i)
		_, _ = fmt.Printf("%-06d %-38v \033[32m%6s\033[0m.\033[36m%-6s\033[0m \033[31m%s\033[0m (%d) %s \n",
			i.LineNr, timestamp, i.Facility, i.Level, i.Entity, i.Pid, i.Message)
		self.printedLineNr = i.LineNr
	}
	matched := (len(parse_config.Entity) == 0 && len(parse_config.Level) == 0 && len(parse_config.Facility) == 0) ||
		len(parse_config.Entity) != 0 && info.Entity == parse_config.Entity ||
		len(parse_config.Level) != 0 && info.Level == parse_config.Level ||
		len(parse_config.Facility) != 0 && info.Facility == parse_config.Facility

	if matched {
		self.selectLineNr = info.LineNr
	}
	if matched && self.Fifo.Fill > 0 {
		// handles the grep BEFORE case
		for {
			v, err := self.Fifo.Get()
			if err != nil {
				break
			}
			printer(v)
		}
	}
	if matched {
		printer(info)
	} else {
		if parse_config.Grep.After > 0 && self.selectLineNr >= 0 && (info.LineNr-self.selectLineNr) <= parse_config.Grep.After {
			// handles the grep AFTER case
			printer(info)
		} else {
			// when not printing just fill the fifo
			if self.Fifo.Cap > 0 {
				self.Fifo.Push(info)
			}
		}
	}
}

func (self *SvLogger) displayTimestamp(i types.Info) string {
	switch self.ParseConfig.TimeConfig {
	case "uptime_s":
		seconds := i.Timestamp.Sub(self.bootTime).Seconds()
		return fmt.Sprintf("%7.03fs", seconds)
		break
	case "local":
		location, _ := time.LoadLocation("Europe/Madrid")
		return i.Timestamp.In(location).Format(time.RFC3339)
		break
	}
	return i.Timestamp.Format(time.RFC3339)
}
