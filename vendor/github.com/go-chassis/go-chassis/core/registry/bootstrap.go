package registry

import (
	"errors"

	"github.com/go-chassis/go-chassis/core/common"
	"github.com/go-chassis/go-chassis/core/config"
	"github.com/go-chassis/go-chassis/core/config/schema"
	"github.com/go-chassis/go-chassis/core/lager"
	"github.com/go-chassis/go-chassis/core/metadata"
	"github.com/go-chassis/go-chassis/pkg/runtime"
)

var errEmptyServiceIDFromRegistry = errors.New("got empty serviceID from registry")

// microServiceDependencies micro-service dependencies
var microServiceDependencies *MicroServiceDependency

// InstanceEndpoints instance endpoints
var InstanceEndpoints map[string]string

// RegisterMicroservice register micro-service
func RegisterMicroservice() error {
	service := config.MicroserviceDefinition
	if e := service.ServiceDescription.Environment; e != "" {
		lager.Logger.Infof("Microservice environment: [%s]", e)
	} else {
		lager.Logger.Debug("No microservice environment defined")
	}
	microServiceDependencies = &MicroServiceDependency{}
	schemas, err := schema.GetSchemaIDs(service.ServiceDescription.Name)
	if err != nil {
		lager.Logger.Warnf("No schemas file for microservice [%s].", service.ServiceDescription.Name)
		schemas = make([]string, 0)
	}
	if service.ServiceDescription.Level == "" {
		service.ServiceDescription.Level = common.DefaultLevel
	}
	framework := metadata.NewFramework()

	microservice := &MicroService{
		ServiceID:   runtime.ServiceID,
		AppID:       config.GlobalDefinition.AppID,
		ServiceName: service.ServiceDescription.Name,
		Version:     service.ServiceDescription.Version,
		Environment: service.ServiceDescription.Environment,
		Status:      common.DefaultStatus,
		Level:       service.ServiceDescription.Level,
		Schemas:     schemas,
		Framework: &Framework{
			Version: framework.Version,
			Name:    framework.Name,
		},
		RegisterBy: framework.Register,
	}
	lager.Logger.Infof("Framework registered is [ %s:%s ]", framework.Name, framework.Version)
	lager.Logger.Infof("Micro service registered by [ %s ]", framework.Register)

	sid, err := DefaultRegistrator.RegisterService(microservice)
	if err != nil {
		lager.Logger.Errorf("Register [%s] failed: %s", microservice.ServiceName, err)
		return err
	}
	if sid == "" {
		lager.Logger.Error(errEmptyServiceIDFromRegistry.Error())
		return errEmptyServiceIDFromRegistry
	}
	runtime.ServiceID = sid
	lager.Logger.Infof("Register [%s/%s] success", runtime.ServiceID, microservice.ServiceName)

	for _, schemaID := range schemas {
		schemaInfo := schema.DefaultSchemaIDsMap[schemaID]
		DefaultRegistrator.AddSchemas(sid, schemaID, schemaInfo)
	}
	if service.ServiceDescription.Properties == nil {
		service.ServiceDescription.Properties = make(map[string]string)
	}

	//update metadata
	if config.GetRegistratorScope() == common.ScopeFull {
		service.ServiceDescription.Properties["allowCrossApp"] = "true"
	} else {
		service.ServiceDescription.Properties["allowCrossApp"] = "false"
	}
	if err := DefaultRegistrator.UpdateMicroServiceProperties(sid, service.ServiceDescription.Properties); err != nil {
		lager.Logger.Errorf("Update micro service properties failed, serviceID = %s. err %s", sid, err)
		return err
	}
	lager.Logger.Debugf("Update micro service properties success, serviceID = %s.", sid)

	return refreshDependency(microservice)
}

// refreshDependency refresh dependency
func refreshDependency(service *MicroService) error {
	providersDependencyMicroService := make([]*MicroService, 0)
	if len(config.GlobalDefinition.Cse.References) == 0 {
		lager.Logger.Info("Don't need add dependency")
		return nil
	}
	for k, v := range config.GlobalDefinition.Cse.References {
		providerDependencyMicroService := &MicroService{
			AppID:       config.GlobalDefinition.AppID,
			ServiceName: k,
			Version:     v.Version,
		}
		providersDependencyMicroService = append(providersDependencyMicroService, providerDependencyMicroService)
	}
	microServiceDependency := &MicroServiceDependency{
		Consumer:  service,
		Providers: providersDependencyMicroService,
	}
	microServiceDependencies = microServiceDependency

	return DefaultRegistrator.AddDependencies(microServiceDependencies)
}

// RegisterMicroserviceInstances register micro-service instances
func RegisterMicroserviceInstances() error {
	lager.Logger.Info("Start to register instance.")
	service := config.MicroserviceDefinition
	var err error

	sid, err := DefaultServiceDiscoveryService.GetMicroServiceID(config.GlobalDefinition.AppID, service.ServiceDescription.Name, service.ServiceDescription.Version, service.ServiceDescription.Environment)
	if err != nil {
		lager.Logger.Errorf("Get service failed, key: %s:%s:%s, err %s",
			config.GlobalDefinition.AppID,
			service.ServiceDescription.Name,
			service.ServiceDescription.Version, err)
		return err
	}
	eps := MakeEndpointMap(config.GlobalDefinition.Cse.Protocols)
	lager.Logger.Infof("service support protocols %s", config.GlobalDefinition.Cse.Protocols)
	if InstanceEndpoints != nil {
		eps = InstanceEndpoints
	}

	microServiceInstance := &MicroServiceInstance{
		EndpointsMap: eps,
		HostName:     runtime.HostName,
		Status:       common.DefaultStatus,
		Metadata:     map[string]string{"nodeIP": config.NodeIP},
	}

	var dInfo = new(DataCenterInfo)
	if config.GlobalDefinition.DataCenter.Name != "" && config.GlobalDefinition.DataCenter.AvailableZone != "" {
		dInfo.Name = config.GlobalDefinition.DataCenter.Name
		dInfo.Region = config.GlobalDefinition.DataCenter.Name
		dInfo.AvailableZone = config.GlobalDefinition.DataCenter.AvailableZone
		microServiceInstance.DataCenterInfo = dInfo
	}

	instanceID, err := DefaultRegistrator.RegisterServiceInstance(sid, microServiceInstance)
	if err != nil {
		lager.Logger.Errorf("Register instance failed, serviceID: %s, err %s", err)
		return err
	}
	//Set to runtime
	runtime.InstanceID = instanceID
	runtime.InstanceStatus = runtime.StatusRunning
	if service.ServiceDescription.InstanceProperties != nil {
		if err := DefaultRegistrator.UpdateMicroServiceInstanceProperties(sid, instanceID, service.ServiceDescription.InstanceProperties); err != nil {
			lager.Logger.Errorf("UpdateMicroServiceInstanceProperties failed, microServiceID/instanceID = %s/%s.", sid, instanceID)
			return err
		}
		lager.Logger.Debugf("UpdateMicroServiceInstanceProperties success, microServiceID/instanceID = %s/%s.", sid, instanceID)
	}

	value, _ := SelfInstancesCache.Get(microServiceInstance.ServiceID)
	instanceIDs, _ := value.([]string)
	var isRepeat bool
	for _, va := range instanceIDs {
		if va == instanceID {
			isRepeat = true
		}
	}
	if !isRepeat {
		instanceIDs = append(instanceIDs, instanceID)
	}
	SelfInstancesCache.Set(sid, instanceIDs, 0)
	lager.Logger.Infof("Register instance success, serviceID/instanceID: %s/%s.", sid, instanceID)
	return nil
}
