package types

import "time"

type Config struct {
	Facilities []string `json:"facilities"`
	Levels     []string `json:"levels"`
	Entities   []string `json:"entities"`
	Services   []string `json:"services"`
}

type Info struct {
	Timestamp time.Time
	Facility  string
	Level     string
	Entity    string
	Pid       int
	Message   string
	LineNr    int
}

type Grep struct {
	After   int
	Before  int
	Context int
}
type ParseConfig struct {
	Facility   string
	Level      string
	Entity     string
	Service    string
	Grep       Grep
	TimeConfig string
	Follow     bool
	Monochrome bool
	AnsiColor  string
}
