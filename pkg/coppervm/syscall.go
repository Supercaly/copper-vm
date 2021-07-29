package coppervm

// Represent a system call on the VM
type SysCall int

const (
	SysCallRead SysCall = iota
	SysCallWrite
)

// Represent a file
type FileDescriptor int

const (
	FdStdIn FileDescriptor = iota
	FdStdOut
	FdStdErr
)
