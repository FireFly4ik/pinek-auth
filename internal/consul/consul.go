package consul

import (
	"auth/internal/config"
	"github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
	"strconv"
	"time"
)

type ConsulProvider struct {
	address string
	client  *api.Client
	name    string
	checkId string
}

func NewProvider(envConf *config.Config) *ConsulProvider {
	client, err := api.NewClient(&api.Config{Address: envConf.Consul.Address})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Consul client")
		return nil
	}

	cp := &ConsulProvider{
		address: envConf.Consul.Address,
		client:  client,
		name:    envConf.Consul.Name,
		checkId: envConf.Consul.CheckId,
	}

	err = cp.registerService(envConf)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register service in Consul")
		return nil
	}

	go cp.updateHealthCheck(envConf)

	return cp
}

func (p *ConsulProvider) registerService(envConf *config.Config) error {
	check := &api.AgentServiceCheck{
		DeregisterCriticalServiceAfter: envConf.Consul.DeregisterTTL,
		TTL:                            envConf.Consul.RegisterTTL,
		TLSSkipVerify:                  true,
		CheckID:                        envConf.Consul.CheckId,
	}

	port, _ := strconv.Atoi(envConf.Port)

	register := &api.AgentServiceRegistration{
		Address: envConf.Address,
		Port:    port,
		ID:      envConf.Consul.Name,
		Name:    envConf.Consul.Name,
		Tags:    []string{"auth"},
		Check:   check,
	}

	err := p.client.Agent().ServiceRegister(register)
	if err != nil {
		return err
	}

	return nil
}

func (p *ConsulProvider) updateHealthCheck(envConf *config.Config) {
	ttl, err := time.ParseDuration(envConf.Consul.RefreshTTL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse TTL duration")
		return
	}

	ticker := time.NewTicker(ttl)

	for {
		err = p.client.Agent().UpdateTTL(p.checkId, "online", api.HealthPassing)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update Consul health check")
		}
		<-ticker.C
	}
}

func (p *ConsulProvider) DeregisterService() {
	err := p.client.Agent().ServiceDeregister(p.name)
	if err != nil {
		log.Error().Err(err).Msg("failed to deregister service from Consul")
	}

	log.Info().Msg("service deregistered from Consul")
}
