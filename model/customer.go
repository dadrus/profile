package model

import (
	"errors"
	"time"
)

var (
	ErrCustomerNotFound = errors.New("Customer not found")
)

var id = 1

type Gender int

const (
	GenderUnknown = 0
	GenderDiverse = 1
	GenderFemale  = 2
	GenderMale    = 3
)

func (g Gender) MarshalJSON() ([]byte, error) {
	switch g {
	case GenderDiverse:
		return []byte("\"diverse\""), nil
	case GenderFemale:
		return []byte("\"female\""), nil
	case GenderMale:
		return []byte("\"male\""), nil
	default:
		return []byte("\"unknown\""), nil
	}
}

func (g *Gender) UnmarshalJSON(raw []byte) error {
	val := string(raw)
	switch val {
	case "\"diverse\"", "\"d\"":
		*g = GenderDiverse
	case "\"female\"", "\"f\"":
		*g = GenderFemale
	case "\"male\"", "\"m\"":
		*g = GenderMale
	default:
		*g = GenderUnknown
	}
	return nil
}

type Address struct {
	Street  string `form:"street_address" json:"street,omitempty"`
	City    string `form:"city" json:"city,omitempty"`
	Zip     string `form:"zip" json:"zip,omitempty"`
	State   string `form:"state" json:"state,omitempty"`
	Country string `form:"country" json:"country,omitempty"`
}

type Customer struct {
	ID          int        `form:"id" json:"id"`
	FirstName   string     `form:"first_name" json:"first_name,omitempty"`
	LastName    string     `form:"last_name" json:"last_name,omitempty"`
	UserName    string     `form:"user_name" json:"user_name,omitempty"`
	Gender      Gender     `form:"gender" json:"gender,omitempty"`
	Birthday    *time.Time `form:"birthday" time_format:"2006-01-02" json:"birthday,omitempty"`
	Address     *Address   `json:"address,omitempty"`
	Email       string     `form:"email" json:"email"`
	PhoneNumber string     `form:"phone" json:"phone,omitempty"`
	Password    string     `form:"-" json:"-"`
}

var customers []*Customer

func NewCustomer(email, password string) *Customer {
	id++

	c := &Customer{
		ID:       id,
		Email:    email,
		Password: password,
		UserName: email,
	}

	customers = append(customers, c)
	return c
}

func FindCustomerById(id int) (*Customer, error) {
	for _, c := range customers {
		if c.ID == id {
			return c, nil
		}
	}

	return nil, ErrCustomerNotFound
}

func FindCustomerByEmail(email string) (*Customer, error) {
	for _, c := range customers {
		if c.Email == email {
			return c, nil
		}
	}

	return nil, ErrCustomerNotFound
}

func FindCustomerByUserName(userName string) (*Customer, error) {
	for _, c := range customers {
		if c.UserName == userName {
			return c, nil
		}
	}

	return nil, ErrCustomerNotFound
}

func CustomerExistsForEmail(email string) bool {
	if _, err := FindCustomerByEmail(email); err == ErrCustomerNotFound {
		return false
	}
	return true
}
