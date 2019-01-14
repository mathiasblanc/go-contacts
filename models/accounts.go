package models

import (
	"fmt"
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	u "github.com/mathiasblanc/go-contacts/utils"
	"golang.org/x/crypto/bcrypt"
)

// Token JWT claims
type Token struct {
	UserID   uint
	Username string
	jwt.StandardClaims
}

// Account A user's account
type Account struct {
	gorm.Model
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token" ;sql:"-"`
}

// Validate validates an account
func (account *Account) Validate() (map[string]interface{}, bool) {
	if !strings.Contains(account.Email, "@") {
		return u.Message(false, "Email address is required"), false
	}

	fmt.Println(account.Password)

	if len(account.Password) < 6 {
		return u.Message(false, "Password must be at least 6 characters long"), false
	}

	temp := &Account{}

	err := GetDB().Table("accounts").Where("email = ?", account.Email).First(temp).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return u.Message(false, "Connection error. Please retry"), false
	}

	if temp.Email != "" {
		return u.Message(false, "Email address already used"), false
	}

	return u.Message(false, "Requirement passed"), true
}

// Create creates an account
func (account *Account) Create() map[string]interface{} {
	if resp, ok := account.Validate(); !ok {
		return resp
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)

	GetDB().Create(account)

	if account.ID <= 0 {
		return u.Message(false, "Failed to create account, connection error")
	}

	tk := &Token{UserID: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString
	account.Password = ""

	response := u.Message(true, "Account has been created")
	response["account"] = account
	return response
}

//Login Logs a user in
func Login(email, password string) map[string]interface{} {
	account := &Account{}
	err := GetDB().Table("accounts").Where("email = ?", email).First(account).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return u.Message(false, "Email address not found")
		}

		return u.Message(false, "Connection error, please retry")
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))

	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return u.Message(false, "Invalid login credentials")
	}

	account.Password = ""

	tk := &Token{UserID: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString

	res := u.Message(true, "Logged In")
	res["account"] = account
	return res
}

// GetUser returns a user from an account
func GetUser(u uint) *Account {
	acc := &Account{}

	GetDB().Table("accounts").Where("id = ?", u).First(acc)

	if acc.Email == "" {
		return nil
	}

	acc.Password = ""
	return acc
}
