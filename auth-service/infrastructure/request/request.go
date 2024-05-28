package request

import "errors"

type Validator interface {
	AreValidRequestData() error
}

type PasswordMatcher interface {
	arePasswordsMatch() error
}

func checkPasswordsMatch(password, confirmPassword string) error {
	if password != confirmPassword {
		return errors.New("passwords do not match")
	}
	return nil
}
