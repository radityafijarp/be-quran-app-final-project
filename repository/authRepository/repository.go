package authRepository

type LoggedInUser struct {
	Username string
}

type Repository struct {
	LoggedInUser *LoggedInUser
}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) Login(username string) error {
	r.LoggedInUser = &LoggedInUser{
		Username: username,
	}

	return nil
}

func (r *Repository) Logout() {
	r.LoggedInUser = nil
}

func (r *Repository) IsLoggedIn() bool {
	return r.LoggedInUser != nil
}
