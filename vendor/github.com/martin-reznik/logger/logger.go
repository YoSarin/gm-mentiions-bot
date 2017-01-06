package logger

import (
	"fmt"
	"github.com/fatih/color"
	"runtime"
	"sync"
	"time"
)

// Log - logger struct
type Log struct {
	LogStream   chan *LogLine
	processor   func(*LogLine)
	wg          *sync.WaitGroup
	LogSeverity map[Severity]bool
}

// Severity - type for severity
type Severity string

const (
	// ERROR - error severity
	ERROR = Severity("error")
	// WARNING - warning severity
	WARNING = Severity("warning")
	// INFO - info severity
	INFO = Severity("info")
	// NOTICE - notice severity
	NOTICE = Severity("notice")
	// DEBUG - debug severity
	DEBUG = Severity("debug")
)

var colorMap = map[Severity]func(...interface{}) string{
	ERROR:   color.New(color.FgRed).SprintFunc(),
	WARNING: color.New(color.FgYellow).SprintFunc(),
	INFO:    color.New(color.FgGreen).SprintFunc(),
	NOTICE:  color.New(color.FgCyan).SprintFunc(),
	DEBUG:   color.New(color.FgBlue).SprintFunc(),
}

// ColoredString - will output severity as a nice colorfull string
func (s *Severity) ColoredString() string {
	return colorMap[*s](string(*s))
}

// LogLine - struct containing info about log
type LogLine struct {
	Message  string
	Severity Severity
	Time     time.Time
	File     string
}

// Config - logger config
type Config struct {
	GoRoutinesLogTicker time.Duration
}

func (c *Config) merge(changes *Config) *Config {
	if changes.GoRoutinesLogTicker > 0 {
		c.GoRoutinesLogTicker = changes.GoRoutinesLogTicker
	}
	return c
}

var defaultConf = Config{
	GoRoutinesLogTicker: 0 * time.Second,
}

// Print - will print logline to stdout
func (l *LogLine) Print() {
	fmt.Printf("[%v] %v \"%v\"\n", l.Severity.ColoredString(), l.Message, l.File)
}

// NewLog - creates new logger
func NewLog(processor func(line *LogLine), conf *Config) *Log {
	c := defaultConf.merge(conf)
	stream := make(chan *LogLine)
	wg := &sync.WaitGroup{}

	l := &Log{
		LogStream:   stream,
		LogSeverity: map[Severity]bool{INFO: true, ERROR: true, WARNING: true, NOTICE: true, DEBUG: false},
		processor:   processor,
		wg:          wg,
	}

	go func(l *Log) {

		if c.GoRoutinesLogTicker <= 0 {
			return
		}

		ticker := time.NewTicker(c.GoRoutinesLogTicker)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.Debug(fmt.Sprintf("Goroutines count: %v", runtime.NumGoroutine()))
			}
		}
	}(l)

	wg.Add(1)
	go func(stream chan *LogLine, processor func(line *LogLine), wg *sync.WaitGroup) {
		defer wg.Done()
		for line := range stream {
			processor(line)
		}
	}(stream, processor, wg)

	return l
}

func (l *Log) log(severity Severity, m string, values ...interface{}) {
	if l.LogSeverity[severity] {
		_, filename, line, _ := runtime.Caller(2)
		message := fmt.Sprintf(m, values...)
		l.LogStream <- &LogLine{
			message,
			severity,
			time.Now(),
			fmt.Sprintf("%v:%v", filename, line),
		}
	}
}

// Notice - puts notice into chan
func (l *Log) Notice(m string, values ...interface{}) {
	l.log(NOTICE, m, values...)
}

// Error - puts error into chan
func (l *Log) Error(m string, values ...interface{}) {
	l.log(ERROR, m, values...)
}

// Info - puts info into chan
func (l *Log) Info(m string, values ...interface{}) {
	l.log(INFO, m, values...)
}

// Warning - puts warning into chan
func (l *Log) Warning(m string, values ...interface{}) {
	l.log(WARNING, m, values...)
}

// Debug - puts debug into chan
func (l *Log) Debug(m string, values ...interface{}) {
	l.log(DEBUG, m, values...)
}

// Close - will close log and wait for log processor to finish
func (l *Log) Close() {
	close(l.LogStream)
	l.wg.Wait()
}
