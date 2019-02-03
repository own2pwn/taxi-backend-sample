package errorswrapper

import (
	"fmt"

	"github.com/pkg/errors"
)

// WrapErrorSlice - собирает сообщения об ошибках из всего ошибок в slice и создает из них один error
func WrapErrorSlice(errs []error) error {
	message := ""
	for _, err := range errs {
		if err == nil {
			continue
		}
		message = fmt.Sprintf("%v\n%v", message, err.Error())
	}
	if message == "" {
		return nil
	}
	return errors.New(message)
}
