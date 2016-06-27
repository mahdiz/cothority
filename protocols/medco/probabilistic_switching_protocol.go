package medco

import (
	"errors"
	"github.com/dedis/cothority/log"
	"github.com/dedis/cothority/network"
	"github.com/dedis/cothority/sda"
	. "github.com/dedis/cothority/services/medco/libmedco"
	"github.com/dedis/crypto/abstract"
)

const PROBABILISTIC_SWITCHING_PROTOCOL_NAME = "ProbabilisticSwitching"

func init() {
	network.RegisterMessageType(ProbabilisticSwitchedMessage{})
	sda.ProtocolRegisterName(PROBABILISTIC_SWITCHING_PROTOCOL_NAME, NewProbabilisticSwitchingProtocol)
}

type ProbabilisticSwitchedMessage struct {
	Data            map[TempID]CipherVector
	TargetPublicKey abstract.Point
	Proof           map[TempID][]CompleteProof
}

type ProbabilisticSwitchedStruct struct {
	*sda.TreeNode
	ProbabilisticSwitchedMessage
}

type ProbabilisticSwitchingProtocol struct {
	*sda.TreeNodeInstance

	// Protocol feedback channel
	FeedbackChannel chan map[TempID]CipherVector

	// Protocol communication channels
	PreviousNodeInPathChannel chan ProbabilisticSwitchedStruct

	// Protocol state data
	nextNodeInCircuit *sda.TreeNode
	TargetOfSwitch    *map[TempID]DeterministCipherVector
	SurveyPHKey       *abstract.Scalar
	TargetPublicKey   *abstract.Point
}

func NewProbabilisticSwitchingProtocol(n *sda.TreeNodeInstance) (sda.ProtocolInstance, error) {
	probabilisticSwitchingProtocol := &ProbabilisticSwitchingProtocol{
		TreeNodeInstance: n,
		FeedbackChannel:  make(chan map[TempID]CipherVector),
	}

	if err := probabilisticSwitchingProtocol.RegisterChannel(&probabilisticSwitchingProtocol.PreviousNodeInPathChannel); err != nil {
		return nil, errors.New("couldn't register data reference channel: " + err.Error())
	}

	var i int
	var node *sda.TreeNode
	var nodeList = n.Tree().List()
	for i, node = range nodeList {
		if n.TreeNode().Equal(node) {
			probabilisticSwitchingProtocol.nextNodeInCircuit = nodeList[(i+1)%len(nodeList)]
			break
		}
	}

	return probabilisticSwitchingProtocol, nil
}

// Starts the protocol
func (p *ProbabilisticSwitchingProtocol) Start() error {

	if p.TargetOfSwitch == nil {
		return errors.New("No map given as probabilistic switching target.")
	}
	if p.TargetPublicKey == nil {
		return errors.New("No map given as target public key.")
	}
	if p.SurveyPHKey == nil {
		return errors.New("No PH key given.")
	}

	log.Lvl1(p.ServerIdentity(), "started a Probabilistic Switching Protocol")

	targetOfSwitch := make(map[TempID]CipherVector, len(*p.TargetOfSwitch))
	for k := range *p.TargetOfSwitch {
		targetOfSwitch[k] = make(CipherVector, len((*p.TargetOfSwitch)[k]))
		for i, dc := range (*p.TargetOfSwitch)[k] {
			var pc CipherText
			pc.K = network.Suite.Point().Null()
			pc.C = dc.Point
			targetOfSwitch[k][i] = pc
		}
	}
	p.sendToNext(&ProbabilisticSwitchedMessage{targetOfSwitch, *p.TargetPublicKey,
		map[TempID][]CompleteProof{}})

	return nil
}

// Dispatch is an infinite loop to handle messages from channels
func (p *ProbabilisticSwitchingProtocol) Dispatch() error {

	probabilisticSwitchingTarget := <-p.PreviousNodeInPathChannel

	phContrib := suite.Point().Mul(suite.Point().Base(), *p.SurveyPHKey)
	for k, v := range probabilisticSwitchingTarget.Data {
		v.ProbabilisticSwitching(&v, phContrib, probabilisticSwitchingTarget.TargetPublicKey)
		probabilisticSwitchingTarget.Data[k] = v
	}

	if p.IsRoot() {
		log.Lvl1(p.ServerIdentity(), "completed probabilistic switching.")
		p.FeedbackChannel <- probabilisticSwitchingTarget.Data
	} else {
		log.Lvl1(p.ServerIdentity(), "carried on probabilistic switching.")
		p.sendToNext(&probabilisticSwitchingTarget.ProbabilisticSwitchedMessage)
	}

	return nil
}

// Sends the message msg to the next node in the circuit based on the next TreeNode in Tree.List() If not visited yet.
// If the message already visited the next node, doesn't send and returns false. Otherwise, return true.
func (p *ProbabilisticSwitchingProtocol) sendToNext(msg interface{}) {
	err := p.SendTo(p.nextNodeInCircuit, msg)
	if err != nil {
		log.Lvl1("Had an error sending a message: ", err)
	}
}
