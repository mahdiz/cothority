package sda2

type TreeNodeInstance interface {
	SendTo(*TreeNode, interface{}) error

	IsLeaf() bool

	IsRoot() bool

	// etc...
}
