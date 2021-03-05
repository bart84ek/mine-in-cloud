package cloud

type Instance struct {
	Id            string
	ReservationId string
	PublicIP      string
	State         string
	Tags          map[string]string
}
