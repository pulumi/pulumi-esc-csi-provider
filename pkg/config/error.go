package config

type ConfigError struct {
	Path string
	Err  error
}

func NewConfigError(path string, err error) *ConfigError {
	return &ConfigError{
		Path: path,
		Err:  err,
	}
}

func (e *ConfigError) Error() string {
	child := e.Err
	if child, ok := child.(*ConfigError); ok {
		return e.Path + "." + child.Error()
	}
	return e.Path + ": " + child.Error()
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}
