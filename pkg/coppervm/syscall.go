package coppervm

// Represent a system call on the VM
type SysCall int

const (
	SysCallRead SysCall = iota
	SysCallWrite
	SysCallOpen
	SysCallClose
	SysCallSeek
)
