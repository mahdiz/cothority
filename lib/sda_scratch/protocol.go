package sda2

// ProtocolInstance is running Protocol that can receive/ send messages directly
// using the Overlay structure => Tree. It does so by using a TreeNodeInstance
// which is connected to the Overlay.
type ProtocolInstance interface {
	TreeNodeInstance

	Processor

	Shutdown()
}
