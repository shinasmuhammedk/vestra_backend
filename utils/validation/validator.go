package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v *validator.Validate

func Init() {
	v = validator.New()
}
func Validate(obj interface{}) error {
	err := v.Struct(obj)
	if err == nil {
		return nil
	}

	if invalidErr, ok := err.(*validator.InvalidValidationError); ok {
		return fmt.Errorf("message:validation failed: %s is invalid or missing", invalidErr.Error())
	}

	if errs, ok := err.(validator.ValidationErrors); ok {
		if errs == nil || len(errs) == 0 {
			return errors.New("message:validation failed but no specific errors were returned")
		}
		structName := getType(obj)
		var errorMessages []string
		for _, e := range errs {
			field := strings.ReplaceAll(e.Namespace(), structName+".", "")
			errorMessages = append(errorMessages, fmt.Sprintf("%s is invalid or missing", field))
		}
		return fmt.Errorf("message:%s", strings.Join(errorMessages, "; "))
	}

	// If it's neither InvalidValidationError nor ValidationErrors, return the error as is
	return fmt.Errorf("message:%s", err.Error())
}

func ValidateVariable(obj interface{}, tags, parameterName string) error {
	err := v.Var(obj, tags)
	if err == nil {
		return nil
	}
	message := parameterName + " is invalid or missing"
	return errors.New(fmt.Sprintf("message:%s", message))
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}