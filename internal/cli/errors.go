package cli

type ExitError struct {
	Code    int
	Message string
}

func (e ExitError) Error() string { return e.Message }

func usage(msg string) error              { return ExitError{Code: 2, Message: msg} }
func configErr(msg string) error          { return ExitError{Code: 3, Message: msg} }
func taskNotFound(msg string) error       { return ExitError{Code: 4, Message: msg} }
func invalidTransition(msg string) error  { return ExitError{Code: 5, Message: msg} }
func artifactValidation(msg string) error { return ExitError{Code: 11, Message: msg} }
func lockConflict(msg string) error       { return ExitError{Code: 12, Message: msg} }
