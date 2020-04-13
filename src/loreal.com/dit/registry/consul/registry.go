package consul

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

// Consul - consul backend struct
type Consul struct {
	Addr          string
	Scheme        string
	Token         string
	KVPath        string
	TagPrefix     string
	Register      bool
	ServiceAddr   string
	ServiceName   string
	ServiceTags   []string
	ServiceStatus []string
	CheckInterval time.Duration
	CheckTimeout  time.Duration
}

// Registry is an implementation of a Registry interface for consul.
type Registry struct {
	client  *api.Client
	DC      string
	Backend *Consul
}

// NewDefaultRegistry - new default consul registry
func NewDefaultRegistry() (*Registry, error) {
	return NewRegistry(&Consul{
		Addr:          "localhost:8500",
		Scheme:        "http",
		KVPath:        "/ceh/config",
		TagPrefix:     "urlprefix-",
		Register:      true,
		ServiceAddr:   ":9998",
		ServiceName:   "ceh",
		ServiceStatus: []string{"passing"},
		CheckInterval: time.Second,
		CheckTimeout:  3 * time.Second,
	})
}

// NewRegistry - create new consul backend
func NewRegistry(backend *Consul) (*Registry, error) {
	// create a reusable client
	c, err := api.NewClient(&api.Config{Address: backend.Addr, Scheme: backend.Scheme, Token: backend.Token})
	if err != nil {
		return nil, err
	}

	// ping the agent
	dc, err := datacenter(c)
	if err != nil {
		log.Printf("[ERR] - [Consul]: Error connecting to consul %q", backend.Addr)
		return nil, err
	}

	// we're good
	log.Printf("[INFO] - [Consul]: Connecting to %q in datacenter %q", backend.Addr, dc)
	return &Registry{client: c, DC: dc, Backend: backend}, nil
}

//Register - implements Registry.Register interface
func (r *Registry) Register(Name, Description, Addr string, Tags ...string) (unRegisterSignal chan bool) {
	var serviceID string
	var serviceRegistered bool
	service, err := serviceRegistration(Addr, Name, Tags, time.Second*10, time.Second*60)
	if err != nil {
		log.Println("[ERR] - [Consul]: Not registering service in consul: ", err)
		return nil
	}

	registered := func() bool {
		if serviceRegistered {
			return true
		}
		if serviceID == "" {
			return false
		}
		services, err := r.client.Agent().Services()
		if err != nil {
			log.Printf("[ERR] - [Consul]: Cannot get service list. %s", err)
			return false
		}
		serviceRegistered = services[serviceID] != nil
		return serviceRegistered
	}

	register := func() {
		if err := r.client.Agent().ServiceRegister(service); err != nil {
			log.Printf("[ERR] - [Consul]: Cannot register service in consul. %s", err)
			return
		}

		log.Printf("[INFO] - [Consul]: Registered service with id %q", service.ID)
		log.Printf("[INFO] - [Consul]: Registered service with address %q", service.Address)
		log.Printf("[INFO] - [Consul]: Registered service with tags %q", strings.Join(service.Tags, ","))
		log.Printf("[INFO] - [Consul]: Registered service with health check to %q", service.Check.HTTP)

		serviceID = service.ID
	}

	unregister := func() {
		r.client.Agent().ServiceDeregister(serviceID)
		log.Println("[INFO] - [Consul]: Service unregistered", serviceID)
	}

	unRegisterSignal = make(chan bool)
	go func() {
		register()
		for {
			select {
			case <-unRegisterSignal:
				unregister()
				unRegisterSignal <- true
				return
			case <-time.After(time.Second):
				if !registered() {
					register()
				}
			}
		}
	}()
	return unRegisterSignal
}

//Unregister - implements Registry.Unregister interface
func (r *Registry) Unregister(unRegisterSignal chan bool) {
	if unRegisterSignal != nil {
		unRegisterSignal <- true // trigger deregistration
		<-unRegisterSignal       // wait for completion
	}
	return
}

// datacenter returns the datacenter of the local agent
func datacenter(c *api.Client) (string, error) {
	self, err := c.Agent().Self()
	if err != nil {
		return "", err
	}

	cfg, ok := self["Config"]
	if !ok {
		return "", errors.New("[Consul]: self.Config not found")
	}
	dc, ok := cfg["Datacenter"].(string)
	if !ok {
		return "", errors.New("[Consul]: self.Datacenter not found")
	}
	return dc, nil
}

func serviceRegistration(addr, name string, tags []string, interval, timeout time.Duration) (*api.AgentServiceRegistration, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	ipstr, portstr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portstr)
	if err != nil {
		return nil, err
	}

	address := hostname
	ip := net.ParseIP(ipstr)
	if ip != nil {
		address = ip.String()
	}

	serviceID := fmt.Sprintf("%s-%s-%d", name, hostname, port)

	checkURL := fmt.Sprintf("http://%s:%d/health", address, port)

	service := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    name,
		Address: address,
		Port:    port,
		Tags:    tags,
		Check: &api.AgentServiceCheck{
			HTTP:     checkURL,
			Interval: interval.String(),
			Timeout:  timeout.String(),
		},
	}
	return service, nil
}
