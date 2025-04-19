package cuserr

type NotFound struct{}

func (e *NotFound) Error() string {
	return "not found"
}

type InvalidCred struct{}

func (e *InvalidCred) Error() string {
	return "invalid credentials"
}

type ItemAlreadyExists struct {
	ItemName string
}

func (i ItemAlreadyExists) Error() string {
	return i.ItemName + " already exusts"
}

type InvalidInput struct {
	FieldName string
	Reason    string
}

func (i InvalidInput) Error() string {
	bodyMsg := "invalid because " + i.Reason
	if i.FieldName == "" {
		return bodyMsg
	}
	return i.FieldName + " " + bodyMsg
}
