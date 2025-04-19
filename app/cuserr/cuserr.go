package cuserr

type NotFound struct {
	Item string
}

func (e NotFound) Error() string {
	bodyMsg := "not found"
	if e.Item == "" {
		return bodyMsg
	}

	return e.Item + " not found"
}

type InvalidCred struct {
	Item   string
	Reason string
}

func (e InvalidCred) Error() string {
	if e.Item == "" || e.Reason == "" {
		return "invalid credentials"
	}
	return e.Item + " invalid because " + e.Reason
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
	if i.FieldName == "" || i.Reason == "" {
		return "invalid input"
	}
	return i.FieldName + " invalid because " + i.Reason
}
