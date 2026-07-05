package goose_logger

import (
	"github.com/rs/zerolog"
)

type Logger interface {
	Infof(msg string, args ...interface{})
	Debugf(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

type logger struct {
	l *zerolog.Logger
}

func (l *logger) Infof(msg string, args ...interface{}) {
	l.l.Info().Msgf(msg, args...)
}

func (l *logger) Debugf(msg string, args ...interface{}) {
	l.l.Debug().Msgf(msg, args...)
}

func (l *logger) Errorf(msg string, args ...interface{}) {
	l.l.Error().Msgf(msg, args...)
}

func (l *logger) Fatal(v ...interface{}) {
	l.l.Print(v...)
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	l.l.Fatal().Msgf(format, v...)
}

func (l *logger) Print(v ...interface{}) {
	l.l.Print(v...)
}

func (l *logger) Println(v ...interface{}) {
	l.l.Print(v...)
}

func (l *logger) Printf(format string, v ...interface{}) {
	l.l.Printf(format, v...)
}

func NewGooseLoggerAdapter(l *zerolog.Logger) Logger {
	return &logger{l: l}
}
