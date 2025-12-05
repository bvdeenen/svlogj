package svlog

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"svlogj/pkg/types"
	"sync/atomic"
	"time"
)

func Svlog(c types.ParseConfig) {
	ParseLog(false, HandleInterpretedLine, c)
}

func ParseLog(stop_when_finished bool, line_handler func(types.Info, types.ParseConfig), c types.ParseConfig) {

	line_pattern := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+) ((\w+)\.(\w+):)?(.*).*$`)

	var cmd *exec.Cmd
	if len(c.Service) == 0 {
		cmd = exec.Command("svlogtail")
	} else {
		cmd = exec.Command("svlogtail", c.Service)
	}
	pipe, _ := cmd.StdoutPipe()
	defer pipe.Close()
	cmd.Start()
	scanner := bufio.NewScanner(pipe)
	var running atomic.Bool
	// go routine to Check if the svlogtail has stopped.
	if stop_when_finished {
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
		}
		guessEntityAndPid(&info)
		line_handler(info, c)
		running.Store(true)
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

func HandleInterpretedLine(info types.Info, parse_config types.ParseConfig) {
	if len(parse_config.Entity) != 0 && info.Entity != parse_config.Entity {
		return
	}
	if len(parse_config.Level) != 0 && info.Level != parse_config.Level {
		return
	}
	if len(parse_config.Facility) != 0 && info.Facility != parse_config.Facility {
		return
	}
	_, _ = fmt.Printf("%-38v \033[32m%6s\033[0m.\033[36m%-6s\033[0m \033[31m%s\033[0m (%d) %s \n",
		info.Timestamp, info.Facility, info.Level, info.Entity, info.Pid, info.Message)
}
