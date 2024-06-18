package domain

const (
	ServiceName                 string = "auth-service"
	AuthContextPath             string = "/auth"
	UserContextPath             string = "/user"
	BearerSchema                string = "Bearer "
	SenderEmailAddress          string = "devopszms2024@gmail.com"
	SubjectVerifyUser           string = "Activate your profile."
	EmailHost                   string = "smtp.gmail.com"
	EmailPort                   int    = 587
	HostRole                    string = "host"
	GuestRole                   string = "guest"
	AppPassword                 string = "itfhpxicunwkabza"
	AdminUsername               string = "admin@test.com"
	AdminPassword               string = "test"
	Authorization               string = "Authorization"
	ContentType                 string = "Content-Type"
	JsonContentType             string = "application/json"
	LocationHeader              string = "Location"
	HealthCheckMessage          string = "AUTH SERVICE IS HEALTH"
	InvalidIDErrorMessage       string = "Invalid user ID"
	InvalidGroupErrorMessage    string = "Invalid user group"
	UserNotVerifiedErrorMessage string = "user account is disabled. check email for verification"
)
