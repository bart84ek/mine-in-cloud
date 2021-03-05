package cloud

type Cloud interface {
	GetInstances() ([]Instance, error)
	CreateInstance(string, string, string) (Instance, error)
}
