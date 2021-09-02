package models

type QPUser struct {
	ID        string `db:"id"`
	Email     string `db:"email"`
	Username  string `db:"username"`
	Password  string `db:"password"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type IQPUser interface {
	Count() (int, error)
	FindByID(ID string) (QPUser, error)
	FindByEmail(email string) (QPUser, error)
	Exists(email string) (bool, error)
	Check(email string, password string) (QPUser, error)
	Create(email string, password string) (QPUser, error)
}
