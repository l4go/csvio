package csvio

import (
	"errors"
)

var ErrCancel = errors.New("cancel")
var ErrSyntax = errors.New("CSV syntax error")
var ErrFieldCount = errors.New("wrong number of fields in line")
var ErrFieldSizeLimit = errors.New("Field size limit error")
var ErrConfig = errors.New("Config error")

type Config struct {
	Comma    byte
	Quote    byte
	UseQuote bool
	UseCRLF  bool
}

var DefaultConfig = Config{
	Comma:    ',',
	Quote:    '"',
	UseQuote: true,
	UseCRLF:  true,
}

func (cnf *Config) Check() error {
	crln := map[byte]struct{}{
		'\r': struct{}{},
		'\n': struct{}{},
	}

	if _, ok := crln[cnf.Comma]; ok {
		return ErrConfig
	}
	if _, ok := crln[cnf.Quote]; ok {
		return ErrConfig
	}

	if cnf.UseQuote && cnf.Comma == cnf.Quote {
		return ErrConfig
	}

	return nil
}
