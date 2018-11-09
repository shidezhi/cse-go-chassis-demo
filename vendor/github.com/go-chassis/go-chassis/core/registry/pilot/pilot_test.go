package pilot

import (
	"net/http/httptest"
	"testing"

	"github.com/go-chassis/go-chassis/core/lager"
	"github.com/go-chassis/go-chassis/core/registry"
	"github.com/go-chassis/go-chassis/pkg/util/tags"
	"github.com/stretchr/testify/assert"
)

var (
	s  = httptest.NewServer(&mockPilotHandler{})
	sd registry.ServiceDiscovery
)

func init() {
	lager.Initialize("stdout", "", "", "", true, 0, 0, 0)
	registry.SetNoIndexCache()
	sd = newDiscoveryService(registry.Options{Addrs: []string{s.Listener.Addr().String()}})
}

func TestPilot_RegisterServiceAndInstance(t *testing.T) {
	microservice, err := sd.GetMicroService("a")
	assert.NoError(t, err)
	assert.Equal(t, "a", microservice.ServiceName)

	serviceID, err := sd.GetMicroServiceID("", "a", "", "")
	assert.NoError(t, err)
	assert.Equal(t, "a", serviceID)

	services, err := sd.GetAllMicroServices()
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(services))

	instances, err := sd.GetMicroServiceInstances("", "a")
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(instances))
	assert.Equal(t, "1.1.1.1_80", instances[0].InstanceID)
	assert.Equal(t, "1.1.1.1:80", instances[0].EndpointsMap["rest"])

	instances, err = sd.FindMicroServiceInstances("", "a", utiltags.Tags{})
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(instances))
	assert.Equal(t, "1.1.1.1_80", instances[0].InstanceID)
	assert.Equal(t, "1.1.1.1:80", instances[0].EndpointsMap["rest"])

	sd.Close()
}
