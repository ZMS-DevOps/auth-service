package request

type Validator interface {
	AreValidRequestData() error
}

type PasswordMatcher interface {
	arePasswordsMatch() error
}
