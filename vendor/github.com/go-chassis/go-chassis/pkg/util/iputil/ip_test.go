package iputil_test

import (
	"github.com/go-chassis/go-chassis/core/common"
	"github.com/go-chassis/go-chassis/pkg/util/iputil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetLocapIp(t *testing.T) {
	localip := iputil.GetLocalIP()
	assert.NotNil(t, localip)
}

func TestLocalhost(t *testing.T) {
	assert.NotEmpty(t, iputil.Localhost())
}

func TestDefaultEndpoint4Protocol(t *testing.T) {
	assert.NotEmpty(t, iputil.DefaultEndpoint4Protocol(common.ProtocolRest))
}

func TestDefaultPort4ProtocolRest(t *testing.T) {
	assert.NotEmpty(t, iputil.DefaultPort4Protocol(common.ProtocolRest))
}

func TestDefaultPort4ProtocolHighway(t *testing.T) {
	assert.NotEmpty(t, iputil.DefaultPort4Protocol(common.ProtocolHighway))
}

func TestDefaultPort4ProtocolNone(t *testing.T) {
	assert.NotEmpty(t, iputil.DefaultPort4Protocol("http"))
}
