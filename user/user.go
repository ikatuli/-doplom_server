package user

import(
	"golang.org/x/crypto/bcrypt"
	"fmt"
)

type User struct {
	Login     string
	HashPassword  []byte
}

func FindUser(login string) User{
	var userProfile User
	var err error
	userProfile.Login = "testLogin"
	userProfile.HashPassword, err = bcrypt.GenerateFromPassword([]byte("test"), 14)

	if err != nil {
        fmt.Println(err)
    }

	return userProfile
}

func CheckCredentials(userProfile User, password string) error {
	pw := []byte(password)
	err := bcrypt.CompareHashAndPassword(userProfile.HashPassword,pw)
	return err
}
