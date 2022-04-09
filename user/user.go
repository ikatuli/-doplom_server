package user

import(
	"golang.org/x/crypto/bcrypt"
	"fmt"
	"database/sql" // пакет для работы с sql
	_ "github.com/lib/pq" // пакет драйвера PostgreSQL
)

type User struct {
	Login     string
	HashPassword  []byte
}

func FindUser(db *sql.DB,login string) User{
	var userProfile User
	//Запрашиваем хеш аккаунта
	rows, err:=db.Query("SELECT password FROM users WHERE login LIKE $1;",login)
	if err != nil {
		fmt.Println(err)
	}

	//Вытаскиваем из запроса хеш
	var password []byte
	rows.Next()
	if err = rows.Scan(&password); err != nil {
			fmt.Println(err)
		}

	userProfile.Login = login
	userProfile.HashPassword = password

	return userProfile
}

func СreateUser(db *sql.DB, login string, passwd string) error {
	var userProfile User
	var err error

	userProfile.Login = login
	userProfile.HashPassword, err = bcrypt.GenerateFromPassword([]byte(passwd), 14) //Генерируем пароль

	if err != nil {
        fmt.Println(err)
    }
	//Записываем пароль в базу данных
	_, err = db.Exec("INSERT INTO users (login,password) VALUES ($1,$2);",userProfile.Login,userProfile.HashPassword)
	
	if err != nil {
        fmt.Println(err)
    }

	return err
}

func ChangeUser(db *sql.DB, login string, passwd string) error {
	var userProfile User
	var err error

	userProfile.Login = login
	userProfile.HashPassword, err = bcrypt.GenerateFromPassword([]byte(passwd), 14) //Генерируем пароль

	if err != nil {
        fmt.Println(err)
    }
	//Записываем пароль в базу данных
	_, err = db.Exec("UPDATE users SET password = $2 WHERE login = $1;",userProfile.Login,userProfile.HashPassword)

	if err != nil {
        fmt.Println(err)
    }

	return err
}


func CheckCredentials(userProfile User, password string) error { //Сопоставляем пароль и хеш
	pw := []byte(password)
	err := bcrypt.CompareHashAndPassword(userProfile.HashPassword,pw)
	return err
}
