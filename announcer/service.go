package announcer

import (
	"github.com/hashicorp/consul/api"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	Datacenter string
	Service    string
	ServiceID  string
	Address    string
	Port       int
	Tags       []string
}

type Announcer struct {
	config Config
}

func New(cfg Config) (a *Announcer) {
	a = &Announcer{
		config: cfg,
	}

	return
}

func (a *Announcer) Register() {
	client, _ := api.NewClient(api.DefaultConfig())
	catalog := client.Catalog()

	service := &api.AgentService{
		ID:      a.config.ServiceID,
		Service: a.config.Service,
		Tags:    a.config.Tags,
		Port:    a.config.Port,
	}

	//check := &api.AgentCheck{}

	reg := &api.CatalogRegistration{
		Datacenter: a.config.Datacenter,
		Node:       a.config.ServiceID,
		Address:    a.config.Address,
		Service:    service,
		Check:      nil,
	}

	if _, err := catalog.Register(reg, nil); err != nil {
		log.Fatal(err)
	}

	a.RegisterShutdown()
}

func (a *Announcer) RegisterShutdown() {
	client, _ := api.NewClient(api.DefaultConfig())
	catalog := client.Catalog()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc

		log.Println("Deregistering...")

		dereg := &api.CatalogDeregistration{
			Datacenter: a.config.Datacenter,
			Node:       a.config.ServiceID,
			Address:    a.config.Address,
			ServiceID:  a.config.ServiceID,
			CheckID:    a.config.ServiceID,
		}

		if _, err := catalog.Deregister(dereg, nil); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()
}
