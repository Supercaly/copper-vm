package coppervm

type CoppervmError int

const (
	ErrorOk CoppervmError = iota
	ErrorIllegalInstAccess
	ErrorStackOverflow
	ErrorStackUnderflow
)

func (err CoppervmError) String() string {
	return [...]string{
		"ErrorOk",
		"ErrorIllegalInstAccess",
		"ErrorStackOverflow",
		"ErrorStackUnderflow",
	}[err]
}
