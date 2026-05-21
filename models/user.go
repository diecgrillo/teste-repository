package models

import (
	"errors"
	"regexp"
	"time"
)

const birthdateLayout = "2006-01-02"

type User struct {
	ID        string `json:"id" dynamodbav:"id"`
	Name      string `json:"name" dynamodbav:"name"`
	CPF       string `json:"cpf" dynamodbav:"cpf"`
	Email     string `json:"email" dynamodbav:"email"`
	Birthdate string `json:"birthdate" dynamodbav:"birthdate"`
	Phone     string `json:"phone" dynamodbav:"phone"`
	CreatedAt string `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt string `json:"updated_at" dynamodbav:"updated_at"`
}

type CreateUserInput struct {
	Name      string `json:"name"`
	CPF       string `json:"cpf"`
	Email     string `json:"email"`
	Birthdate string `json:"birthdate"`
	Phone     string `json:"phone"`
}

type UpdateUserInput struct {
	Name      *string `json:"name"`
	Email     *string `json:"email"`
	Birthdate *string `json:"birthdate"`
	Phone     *string `json:"phone"`
}

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	// accepts 11 raw digits or formatted XXX.XXX.XXX-XX
	cpfRegex = regexp.MustCompile(`^(\d{3}\.?\d{3}\.?\d{3}-?\d{2}|\d{11})$`)
	phoneRegex = regexp.MustCompile(`^\+?[\d\s\-().]{7,20}$`)
)

func (i *CreateUserInput) Validate() error {
	if i.Name == "" {
		return errors.New("name is required")
	}
	if i.CPF == "" {
		return errors.New("cpf is required")
	}
	if !cpfRegex.MatchString(i.CPF) {
		return errors.New("cpf must be 11 digits or formatted as XXX.XXX.XXX-XX")
	}
	if !validateCPFDigits(i.CPF) {
		return errors.New("cpf is invalid")
	}
	if i.Email == "" {
		return errors.New("email is required")
	}
	if !emailRegex.MatchString(i.Email) {
		return errors.New("email format is invalid")
	}
	if i.Birthdate == "" {
		return errors.New("birthdate is required")
	}
	if err := validateDate(i.Birthdate); err != nil {
		return err
	}
	if i.Phone == "" {
		return errors.New("phone is required")
	}
	if !phoneRegex.MatchString(i.Phone) {
		return errors.New("phone format is invalid")
	}
	return nil
}

func (i *UpdateUserInput) Validate() error {
	if i.Email != nil {
		if !emailRegex.MatchString(*i.Email) {
			return errors.New("email format is invalid")
		}
	}
	if i.Birthdate != nil {
		if err := validateDate(*i.Birthdate); err != nil {
			return err
		}
	}
	if i.Phone != nil {
		if !phoneRegex.MatchString(*i.Phone) {
			return errors.New("phone format is invalid")
		}
	}
	return nil
}

func validateDate(date string) error {
	t, err := time.Parse(birthdateLayout, date)
	if err != nil {
		return errors.New("birthdate must be in YYYY-MM-DD format")
	}
	if t.After(time.Now()) {
		return errors.New("birthdate cannot be in the future")
	}
	return nil
}

// validateCPFDigits checks the two check digits of a Brazilian CPF.
func validateCPFDigits(cpf string) bool {
	digits := regexp.MustCompile(`\D`).ReplaceAllString(cpf, "")
	if len(digits) != 11 {
		return false
	}
	// reject all-same sequences like 00000000000
	allSame := true
	for _, c := range digits[1:] {
		if c != rune(digits[0]) {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (10 - i)
	}
	r := sum % 11
	d1 := 0
	if r >= 2 {
		d1 = 11 - r
	}
	if int(digits[9]-'0') != d1 {
		return false
	}

	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(digits[i]-'0') * (11 - i)
	}
	r = sum % 11
	d2 := 0
	if r >= 2 {
		d2 = 11 - r
	}
	return int(digits[10]-'0') == d2
}
