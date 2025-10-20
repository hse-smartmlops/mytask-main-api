package logger

import (
	"log"
	"os"
)

type Logger struct {
    info  *log.Logger
    error *log.Logger
    debug *log.Logger
}

func New() *Logger {
    flags := log.Ldate | log.Ltime | log.Lshortfile
    
    return &Logger{
        info:  log.New(os.Stdout, "INFO: ", flags),
        error: log.New(os.Stderr, "ERROR: ", flags),
        debug: log.New(os.Stdout, "DEBUG: ", flags),
    }
}

func (l *Logger) Info(msg string, args ...any) {
    if len(args) > 0 {
        l.info.Printf(msg, args...)
    } else {
        l.info.Println(msg)
    }
}

func (l *Logger) Error(msg string, args ...any) {
    if len(args) > 0 {
        l.error.Printf(msg, args...)
    } else {
        l.error.Println(msg)
    }
}

func (l *Logger) Debug(msg string, args ...any) {
    if os.Getenv("DEBUG") == "true" {
        if len(args) > 0 {
            l.debug.Printf(msg, args...)
        } else {
            l.debug.Println(msg)
        }
    }
}

func (l *Logger) Fatal(msg string, args ...any) {
    if len(args) > 0 {
        l.error.Fatalf(msg, args...)
    } else {
        l.error.Fatalln(msg)
    }
}