package host

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHostCreation(t *testing.T) {
	h, e := New(HOST_TYPE_SSH)
	assert.NotNil(t, h)
	assert.Nil(t, e)

	h, e = New(HOST_TYPE_DOCKER)
	assert.NotNil(t, h)
	assert.Nil(t, e)

	h, e = New(-1)
	assert.Nil(t, h)
	assert.Contains(t, e.Error(), "host type must be of the HOST_TYPE_{DOCKER,SSH} const")

	h, e = New(24)
	assert.Nil(t, h)
	assert.Contains(t, e.Error(), "host type must be of the HOST_TYPE_{DOCKER,SSH} const")
}

func TestHostTypePredicates(t *testing.T) {
	h, _ := New(HOST_TYPE_SSH)

	assert.True(t, h.IsSshHost())
	assert.False(t, h.IsDockerHost())

	h, _ = New(HOST_TYPE_DOCKER)

	assert.True(t, h.IsDockerHost())
	assert.False(t, h.IsSshHost())
}

func TestIPAddressHandling(t *testing.T) {
	h, e := New(HOST_TYPE_SSH)

	assert.Equal(t, h.GetPublicIPAddress(), "")

	e = h.SetPublicIPAddress("not an IP address")
	assert.Contains(t, e.Error(), "not a valid IP address (either IPv4 or IPv6): not an IP address")

	e = h.SetPublicIPAddress("666.666.666.666")
	assert.Contains(t, e.Error(), "not a valid IP address (either IPv4 or IPv6): 666.666.666.666")

	e = h.SetPublicIPAddress("127.0.0.1")
	assert.Nil(t, e)
	assert.Equal(t, h.GetPublicIPAddress(), "127.0.0.1")
}

func TestUserHandling(t *testing.T) {
	h, _ := New(HOST_TYPE_SSH)

	assert.Equal(t, h.GetUser(), "root")
	assert.Equal(t, h.IsSudoRequired(), false)

	h.SetUser("root")
	assert.Equal(t, h.GetUser(), "root")
	assert.Equal(t, h.IsSudoRequired(), false)

	h.SetUser("gfrey")
	assert.Equal(t, h.GetUser(), "gfrey")
	assert.Equal(t, h.IsSudoRequired(), true)
}
