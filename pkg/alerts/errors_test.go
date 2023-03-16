package alerts

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorsCollection(t *testing.T) {

	var errs Errors
	assert.True(t, errs == nil)

	errs.collect(errors.New("err 1"))
	errs.collect(errors.New("err 2"))

	assert.True(t, errs != nil)

	assert.True(t, len(errs) == 2)

	assert.True(t, errs.Error() == "errors:\nError 0: err 1\nError 1: err 2\n")
}
