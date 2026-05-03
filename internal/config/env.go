package config

import "fmt"

// Env represents the environment type in which the application instance is running.
type Env uint8

const (
	Dev Env = iota
	Prod
)

func (e Env) String() string {
	switch e {
	case Dev:
		return "dev"
	case Prod:
		return "prod"
	default:
		return fmt.Sprintf("undefined: %d", e)
	}
}

func (e *Env) SetValue(s string) error {
	switch s {
	case "dev":
		*e = Dev
		return nil
	case "prod":
		*e = Prod
		return nil
	default:
		return fmt.Errorf("undefined: %s", s)
	}
}
