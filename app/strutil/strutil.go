package strutil

import (
	"wonk/app/cuserr"
)

const (
	MAX_STRING_LENGTH = 32
)

func StrPtr(s string) *string {
	return &s
}

func IsStringValid(s, fieldName string) error {
	if s == "" {
		return cuserr.InvalidInput{FieldName: fieldName, Reason: "value is empty"}
	}
	if len(s) > MAX_STRING_LENGTH {
		return cuserr.InvalidInput{FieldName: fieldName, Reason: "value is too long"}
	}
	return nil
}

func IsPasswordValid(p string) error {
	err := IsStringValid(p, "password")
	if err != nil {
		return err
	}

	return nil
}

func ConvertMonth(monthNum int) string {
	switch monthNum {
	case 1:
		return "January"
	case 2:
		return "February"
	case 3:
		return "March"
	case 4:
		return "April"
	case 5:
		return "May"
	case 6:
		return "June"
	case 7:
		return "July"
	case 8:
		return "August"
	case 9:
		return "September"
	case 10:
		return "October"
	case 11:
		return "November"
	case 12:
		return "December"
	default:
		return "Error"
	}
}
