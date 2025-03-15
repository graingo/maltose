package merror

import "github.com/graingo/maltose/errors/mcode"

// IEqual defines the interface for comparing two errors for equality.
type IEqual interface {
	Error() string
	Equal(target error) bool
}

// ICode defines the interface for the Code functionality.
type ICode interface {
	Error() string
	Code() mcode.Code
}

// IStack defines the interface for the Stack functionality.
type IStack interface {
	Error() string
	Stack() string
}

// ICause defines the interface for the Cause functionality.
type ICause interface {
	Error() string
	Cause() error
}

// ICurrent defines the interface for the Current functionality.
type ICurrent interface {
	Error() string
	Current() error
}

// IUnwrap defines the interface for the Unwrap functionality.
type IUnwrap interface {
	Error() string
	Unwrap() error
}

const (
	commaSeparatorSpace = ", "
)
