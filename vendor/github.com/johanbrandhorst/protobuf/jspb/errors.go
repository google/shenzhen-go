package jspb

import "github.com/gopherjs/gopherjs/js"

// catchException recovers any JS exceptions and
// stores the error in the parameter
func catchException(err *error) {
	e := recover()

	if e == nil {
		return
	}

	if e, ok := e.(*js.Error); ok {
		*err = e
	} else {
		panic(e)
	}
}
