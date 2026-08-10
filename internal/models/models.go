package models

type Result struct {
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
}
