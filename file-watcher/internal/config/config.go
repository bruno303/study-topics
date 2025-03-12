package config

import "os"

type Config struct {
	FilesToWatch []string
	Commands     []string
	Delay        int
	Signal       os.Signal
}

func NewConfig(filesToWatch []string, commands []string, delay int, signal os.Signal) Config {
	return Config{
		FilesToWatch: filesToWatch,
		Commands:     commands,
		Delay:        delay,
		Signal:       signal,
	}
}
