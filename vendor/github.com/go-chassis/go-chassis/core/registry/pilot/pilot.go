package pilot

import (
	"fmt"

	"github.com/go-chassis/go-chassis/core/lager"
	"github.com/go-chassis/go-chassis/core/registry"
	"github.com/go-chassis/go-chassis/pkg/util/tags"
)

// PilotPlugin is the constant string of the plugin name
const PilotPlugin = "pilot"

// Registrator is the struct to do service discovery from istio pilot server
type Registrator struct {
	Name           string
	registryClient *EnvoyDSClient
}

// Close : Close all connection.
func (r *Registrator) Close() error {
	return close(r.registryClient)
}

// ServiceDiscovery is the struct to do service discovery from istio pilot server
type ServiceDiscovery struct {
	Name           string
	registryClient *EnvoyDSClient
}

// GetMicroServiceID : 获取指定微服务的MicroServiceID
func (r *ServiceDiscovery) GetMicroServiceID(appID, microServiceName, version, env string) (string, error) {
	_, err := r.registryClient.GetServiceHosts(microServiceName)
	if err != nil {
		lager.Logger.Errorf("GetMicroServiceID failed: %s", err)
		return "", err
	}
	lager.Logger.Debugf("GetMicroServiceID success")
	return microServiceName, nil
}

// GetAllMicroServices : Get all MicroService information.
func (r *ServiceDiscovery) GetAllMicroServices() ([]*registry.MicroService, error) {
	svcs, err := r.registryClient.GetAllServices()
	if err != nil {
		lager.Logger.Errorf("GetAllMicroServices failed: %s", err)
		return nil, err
	}

	var mss []*registry.MicroService
	for _, s := range svcs {
		mss = append(mss, ToMicroService(s))
	}
	return mss, nil
}

// GetMicroService : 根据microServiceID获取对应的微服务信息
func (r *ServiceDiscovery) GetMicroService(microServiceID string) (*registry.MicroService, error) {
	hs, err := r.registryClient.GetServiceHosts(microServiceID)
	if err != nil {
		lager.Logger.Errorf("GetMicroServiceID failed: %s", err)
		return nil, err
	}
	lager.Logger.Debugf("GetMicroServices success, MicroService: %s", microServiceID)
	return ToMicroService(&Service{
		ServiceKey: microServiceID,
		Hosts:      hs.Hosts,
	}), nil
}

// GetMicroServiceInstances : 获取指定微服务的所有实例
func (r *ServiceDiscovery) GetMicroServiceInstances(consumerID, providerID string) ([]*registry.MicroServiceInstance, error) {
	hs, err := r.registryClient.GetServiceHosts(providerID)
	if err != nil {
		lager.Logger.Errorf("GetMicroServiceInstances failed: %s", err)
		return nil, err
	}
	instances := filterInstances(hs.Hosts)
	lager.Logger.Debugf("GetMicroServiceInstances success, consumerID/providerID: %s/%s", consumerID, providerID)
	return instances, nil
}

// FindMicroServiceInstances find micro-service instances
func (r *ServiceDiscovery) FindMicroServiceInstances(consumerID, microServiceName string, tags utiltags.Tags) ([]*registry.MicroServiceInstance, error) {
	serviceKey := pilotServiceKey(microServiceName)
	value, boo := registry.MicroserviceInstanceIndex.Get(serviceKey, tags.KV)
	if !boo || value == nil {
		lager.Logger.Warnf("%s Get instances from remote, key: %s, %v", consumerID, serviceKey, tags.String())
		hs, err := r.registryClient.GetHostsByKey(serviceKey, tags.KV)
		if err != nil {
			return nil, fmt.Errorf("FindMicroServiceInstances failed, ProviderID: %s, err: %s",
				microServiceName, err)
		}

		filterRestore(hs.Hosts, serviceKey, tags.KV)
		value, boo = registry.MicroserviceInstanceIndex.Get(serviceKey, tags.KV)
		if !boo || value == nil {
			lager.Logger.Debugf("Find no microservice instances for %s from cache", serviceKey)
			return nil, nil
		}
	}
	microServiceInstance, ok := value.([]*registry.MicroServiceInstance)
	if !ok {
		lager.Logger.Errorf("FindMicroServiceInstances failed, Type asserts failed. consumerIDL: %s",
			consumerID)
	}
	return microServiceInstance, nil
}

// AutoSync updating the cache manager
func (r *ServiceDiscovery) AutoSync() {
	c := &CacheManager{
		registryClient: r.registryClient,
	}
	c.AutoSync()
}

// Close : Close all connection.
func (r *ServiceDiscovery) Close() error {
	return close(r.registryClient)
}

func newDiscoveryService(options registry.Options) registry.ServiceDiscovery {
	//TODO: now no tag information can obtain from SDS response
	// tags should rebuild according to RDS requests

	c := &EnvoyDSClient{}
	c.Initialize(Options{
		Addrs:     options.Addrs,
		TLSConfig: options.TLSConfig,
	})
	return &ServiceDiscovery{
		Name:           PilotPlugin,
		registryClient: c,
	}
}

// register pilot registry plugin when import this package
func init() {
	registry.InstallServiceDiscovery(PilotPlugin, newDiscoveryService)
}
