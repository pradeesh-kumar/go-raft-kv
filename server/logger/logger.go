package logger

import (
	"flag"
	"log"
	"os"
	"strconv"
)

const (
	levelInfo  = "INFO "
	levelWarn  = "WARN "
	levelError = "ERROR "
	levelDebug = "DEBUG "
	levelFatal = "FATAL "
)

type CliValue struct {
	IsSet bool
}

func (i *CliValue) String() string {
	return strconv.FormatBool(i.IsSet)
}

func (i *CliValue) Set(_ string) error {
	i.IsSet = true
	return nil
}

var verbose bool

func init() {
	flag.BoolVar(&verbose, "v", false, "Verbose Output")
}

func Info(v ...any) {
	write(levelInfo, v)
}

func Infof(format string, v ...any) {
	writef(levelInfo, format, v)
}

func Warn(v ...any) {
	write(levelWarn, v)
}

func Warnf(format string, v ...any) {
	writef(levelWarn, format, v)
}

func Error(v ...any) {
	write(levelError, v)
}

func Errorf(format string, v ...any) {
	writef(levelError, format, v)
}

func Debug(v ...any) {
	if verbose {
		write(levelDebug, v)
	}
}

func Debugf(format string, v ...any) {
	writef(levelDebug, format, v)
}

func Fatal(v ...any) {
	write(levelFatal, v)
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	writef(levelFatal, format, v)
}

func write(level string, v []any) {
	log.SetPrefix(level)
	log.Println(v...)
}

func writef(level string, format string, v []any) {
	log.SetPrefix(level)
	log.Printf(format, v...)
}
