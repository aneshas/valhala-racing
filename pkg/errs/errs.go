package errs

import "errors"

// ErrTransientPaymentFailure signifies transient payment failure
var ErrTransientPaymentFailure = errors.New("transient error")
