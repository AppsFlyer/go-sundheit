package checks

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	checkName = "xxx"
)

func TestNewPingCheck_nilPinger(t *testing.T) {
	check, err := NewPingCheck(checkName, nil, time.Second)
	assert.Error(t, err, "check creation should fail for nil pinger")
	assert.Nil(t, check, "check creation should fail for nil pinger")
}

func TestNewPingCheck(t *testing.T) {
	assertions := assert.New(t)

	check, err := NewPingCheck(checkName, mockPinger(false), time.Microsecond)
	assertions.NoError(err, "check creation should succeed")
	assertions.NotNil(check, "check creation should succeed")

	assertions.Equal(checkName, check.Name(), "check name")

	_, err = check.Execute(context.Background())
	assertions.NoError(err)

	check, err = NewPingCheck(checkName, mockPinger(true), time.Microsecond)
	assertions.NoError(err, "check creation should succeed")
	assertions.NotNil(check, "check creation should succeed")

	_, err = check.Execute(context.Background())
	assertions.Error(err)
}

func mockPinger(failing bool) PingContextFunc {
	return func(ctx context.Context) error {
		if failing {
			return errors.New("mock fail")
		}

		return nil
	}
}

func TestNewDialPinger(t *testing.T) {
	assertions := assert.New(t)

	pinger := NewDialPinger("tcp", "there.should.be.no.such.host.com:666")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	assertions.Error(pinger.PingContext(ctx), "expecting a ping error for non existing address")

	pinger = NewDialPinger("tcp", "example.com:80")
	assertions.NoError(pinger.PingContext(ctx), "expecting success for an existing address")
}
