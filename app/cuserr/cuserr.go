package cuserr

type NotFound struct{}

func (e *NotFound) Error() string {
	return "not found"
}

type InvalidCred struct{}

func (e *InvalidCred) Error() string {
	return "invalid credentials"
}
