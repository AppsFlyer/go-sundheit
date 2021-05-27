package checks

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	checkName = "xxx"
)

func TestNewPingCheck_nilPinger(t *testing.T) {
	check, err := NewPingCheck(checkName, nil)
	assert.Error(t, err, "check creation should fail for nil pinger")
	assert.Nil(t, check, "check creation should fail for nil pinger")
}

func TestNewPingCheck(t *testing.T) {
	assertions := assert.New(t)

	check, err := NewPingCheck(checkName, mockPinger(false))
	assertions.NoError(err, "check creation should succeed")
	assertions.NotNil(check, "check creation should succeed")

	assertions.Equal(checkName, check.Name(), "check name")

	_, err = check.Execute(context.Background())
	assertions.NoError(err)

	check, err = NewPingCheck(checkName, mockPinger(true))
	assertions.NoError(err, "check creation should succeed")
	assertions.NotNil(check, "check creation should succeed")

	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
	defer cancel()
	_, err = check.Execute(ctx)
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

func TestNewDialPinger_bogusAddress(t *testing.T) {
	assertions := assert.New(t)

	pinger := NewDialPinger("tcp", "there.should.be.no.such.host.com:666")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	assertions.Error(pinger.PingContext(ctx), "expecting a ping error for non existing address")
}

func TestNewDialPinger_existingAddress(t *testing.T) {
	assertions := assert.New(t)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "failed to start a test listener")
	defer func() { _ = ln.Close() }()

	pinger := NewDialPinger("tcp", ln.Addr().String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	assertions.NoError(pinger.PingContext(ctx), "expecting success for an existing address")
}
