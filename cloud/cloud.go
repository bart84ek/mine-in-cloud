package cloud

type Cloud interface {
	GetInstances() ([]Instance, error)
	CreateInstance(string, string, string) (Instance, error)
	GetAddresses()
	Terminate(string) error
	AssignIP(string, string) error
}
