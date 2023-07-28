package tools

import "strings"

type ErrorList struct {
	errors []error
}

func NewErrorList() ErrorList {
	return ErrorList{}
}

func (errList *ErrorList) Append(err error) {
	errList.errors = append(errList.errors, err)
}

func (errList *ErrorList) ToError() error {
	if len(errList.errors) == 0 {
		return nil
	}
	return errList
}

func (errList *ErrorList) Error() string {
	stringBuilder := strings.Builder{}
	for _, err := range errList.errors {
		stringBuilder.WriteString(err.Error())
	}
	return stringBuilder.String()
}
