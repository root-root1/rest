package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/root-root1/rest/internal/validator"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	Id         int64     `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Password   password  `json:"-"`
	Activation bool      `json:"activation"`
	Version    int       `json:"-"`
}

type password struct {
	PlainText *string
	hash      []byte
}

var (
	ErrDuplicateError = errors.New("Duplicate Error")
)

type UserModel struct {
	DB *sql.DB
}

func (p *password) Set(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)

	if err != nil {
		return err
	}

	p.PlainText = &plaintext
	p.hash = hash
	return nil
}

func (p *password) Matches(plaintextpassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextpassword))

	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func validate(v *validator.Validator, email string) {
	v.Check(email != "", "email", "email must provided")
	v.Check(validator.Match(email, validator.EmailRX), "email", "must be a valid email")
}

func ValidatePasswordPlaintext(v *validator.Validator, plaintext string) {
	v.Check(plaintext != "", "password", "Password must be provided")
	v.Check(len(plaintext) >= 8, "password", "must be At least (8 character) long")
	v.Check(len(plaintext) <= 72, "password", "must not be (72 Character) long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "user", "must be provided")
	v.Check(len(user.Name) <= 500, "user", "must not be greater (72 character) long")

	validate(v, user.Email)

	if user.Password.PlainText != nil {
		ValidatePasswordPlaintext(v, *user.Password.PlainText)
	}

	if user.Password.hash != nil {
		panic("missing password user for hash")
	}
}

func (m UserModel) Insert(user *User) error {
	query := `
		insert into users (name, email, password_hash, activation)
		values ($1,$2,$3,$4)
		returning id,created_at, version
	`

	args := []interface{}{user.Name, user.Email, user.Password, user.Activation}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Id, &user.CreatedAt, &user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint \"users_email_key\"`:
			return ErrDuplicateError
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetUserByEmail(email string) (*User, error) {
	query := `
		select id, created_at, name, email, password_hash, activation, version
		from users
		where email = $1
	`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.Id,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activation,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) Update(user *User) error {
	query := `
		update users
		set name = $1, email = $2, password_hash = $3, activation = $4, version = version + 1
		where id = $5 and version = $6
		returning version
	`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activation,
		user.Id,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateError
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
