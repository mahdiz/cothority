package prifi

import (
	"testing"

	"github.com/dedis/cothority/log"
	"github.com/dedis/cothority/sda"
)

func TestMain(m *testing.M) {
	log.MainTest(m)
}

func TestServiceTemplate(t *testing.T) {
	local := sda.NewLocalTest()
	defer local.CloseAll()
	hosts, roster, _ := local.MakeHELS(5, serviceID)
	log.Lvl1("Roster is", roster)

	var services []*Service
	for _, h := range hosts {
		service := local.Services[h.ServerIdentity.ID][serviceID].(*Service)
		services = append(services, service)
	}

	services[0].StartTrustee()
	services[1].StartTrustee()
	services[2].StartRelay()
	services[3].StartClient()
	services[4].StartClient()

	// Now do something with the client and exit
}
