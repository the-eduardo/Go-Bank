package val

import (
	"fmt"
	"net/mail"
	"regexp"
)

var isValidUsername = regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString
var isValidFullName = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString

func ValidateString(value string, minLenght int, maxLength int) error {
	if len(value) < minLenght || len(value) > maxLength {
		return fmt.Errorf("value length must be between %d and %d characters", minLenght, maxLength)
	}
	return nil
}

func ValidateUsername(value string) error {
	if err := ValidateString(value, 3, 30); err != nil {
		return err
	}
	if !isValidUsername(value) {
		return fmt.Errorf("username can only contain letters, numbers and underscores")
	}
	return nil
}
func ValidateFullName(value string) error {
	if err := ValidateString(value, 3, 90); err != nil {
		return err
	}
	if !isValidFullName(value) {
		return fmt.Errorf("names can only contain letters and spaces")
	}
	return nil
}

func ValidatePassword(value string) error {
	if err := ValidateString(value, 6, 72); err != nil {
		return err
	}
	return nil
}

func ValidateEmail(value string) error {
	if err := ValidateString(value, 9, 320); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("invalid email address")
	}
	return nil
}
