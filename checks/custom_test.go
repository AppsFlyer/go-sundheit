package checks

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	chk := CustomCheck{}
	assert.Equal(t, "", chk.Name(), "empty custom check")

	const expectedName = "my.check"
	chk = CustomCheck{CheckName: expectedName}
	assert.Equal(t, expectedName, chk.Name(), "named custom check")
}

func TestExecute(t *testing.T) {
	chk := CustomCheck{}
	details, err := chk.Execute()
	assert.Nil(t, err, "nil check func should execute and return nil error")
	assert.Equal(t, "Unimplemented check", details, "nil check func should execute and return details")

	const expectedDetails = "my.details"
	expectedErr := errors.New("my.error")
	chk.CheckFunc = func() (details interface{}, err error) {
		return expectedDetails, expectedErr
	}

	details, err = chk.Execute()
	assert.Equal(t, expectedDetails, details)
	assert.Equal(t, expectedErr, err)
}
