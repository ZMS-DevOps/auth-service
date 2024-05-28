package domain

const (
	AuthContextPath             string = "/auth"
	UserContextPath             string = "/user"
	BearerSchema                string = "Bearer "
	VerificationEmailTemplate   string = "sendVerificationCode.html"
	SenderEmailAddress          string = "devopszms2024@gmail.com"
	SubjectVerifyUser           string = "Activate your profile."
	EmailHost                   string = "smtp.gmail.com"
	EmailPort                   int    = 587
	AppPassword                 string = "itfhpxicunwkabza"
	AdminUsername               string = "admin@test.com"
	AdminPassword               string = "test"
	Authorization               string = "Authorization"
	ContentType                 string = "Content-Type"
	JsonContentType             string = "application/json"
	LocationHeader              string = "Location"
	HealthCheckMessage          string = "AUTH SERVICE IS HEALTH"
	InvalidIDErrorMessage       string = "Invalid user ID"
	UserNotVerifiedErrorMessage string = "user account is disabled. check email for verification"
)
