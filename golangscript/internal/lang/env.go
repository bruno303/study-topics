package lang

type Environment struct {
	vars map[string]float64
}

func NewEnvironment() *Environment {
	return &Environment{vars: make(map[string]float64)}
}

func (e *Environment) Set(name string, val float64) {
	e.vars[name] = val
}

func (e *Environment) Get(name string) (float64, bool) {
	val, ok := e.vars[name]
	return val, ok
}

func (e *Environment) GetVars() map[string]float64 {
	return e.vars
}
