package longlegs

import "errors"

var (
	ErrNotHTML       = errors.New("Not HTML")
	ErrInvalidURL    = errors.New("Invalid URL")
	ErrRequestFailed = errors.New("RequestFailed")
	ErrNotOk         = errors.New("Response not OK")
	ErrCantReadBody  = errors.New("Can't ready body")
	ErrCantParseHTML = errors.New("Can't parse HTML")
)
