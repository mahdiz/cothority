package example_handlers

import (
	"errors"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/dedis/cothority/lib/dbg"
	"github.com/dedis/cothority/lib/monitor"
	"github.com/dedis/cothority/lib/sda"
)

/*
This is a simple ExampleHandlers-protocol with two steps:
- announcement - which sends a message to all children
- reply - used for counting the number of children
*/

func init() {
	sda.SimulationRegister("ExampleHandlers", NewExampleHandlersSimulation)
}

type ExampleHandlersSimulation struct {
	sda.SimulationBFTree
}

func NewExampleHandlersSimulation(config string) (sda.Simulation, error) {
	es := &ExampleHandlersSimulation{}
	_, err := toml.Decode(config, es)
	if err != nil {
		return nil, err
	}
	return es, nil
}

func (e *ExampleHandlersSimulation) Setup(dir string, hosts []string) (
	*sda.SimulationConfig, error) {
	sc := &sda.SimulationConfig{}
	e.CreateEntityList(sc, hosts, 2000)
	err := e.CreateTree(sc)
	if err != nil {
		return nil, err
	}
	return sc, nil
}

func (e *ExampleHandlersSimulation) Run(config *sda.SimulationConfig) error {
	size := config.Tree.Size()
	dbg.Lvl2("Size is:", size, "rounds:", e.Rounds)
	for round := 0; round < e.Rounds; round++ {
		dbg.Lvl1("Starting round", round)
		round := monitor.NewTimeMeasure("round")
		n, err := config.Overlay.StartNewNodeName("ExampleHandlers", config.Tree)
		if err != nil {
			return err
		}
		children := <-n.ProtocolInstance().(*ProtocolExampleHandlers).ChildCount
		round.Record()
		if children != size {
			return errors.New("Didn't get " + strconv.Itoa(size) +
				" children")
		}
	}
	return nil
}
