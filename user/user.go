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
	Role string
}

var Role = []string{"admin","user"} //Список ролей

func FindUser(db *sql.DB,login string) User{
	var userProfile User
	//Запрашиваем хеш аккаунта
	rows, err:=db.Query("SELECT password,role FROM users WHERE login = $1;",login)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	//Вытаскиваем из запроса хеш
	var password []byte
	var role_int int
	rows.Next()
	if err = rows.Scan(&password,&role_int); err != nil {
			fmt.Println(err)
		}

	userProfile.Login = login
	userProfile.HashPassword = password
	userProfile.Role=Role[role_int]

	return userProfile
}

func СreateUser(db *sql.DB, login string, passwd string, role string) error {
	var role_int int
	var err error

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(passwd), 14) //Генерируем пароль

	for i, v := range Role{ //Поиск номера роли
		if role == v {
			role_int=i
			break
		}
	}

	if err != nil {
        fmt.Println(err)
    }
	//Записываем пароль в базу данных
	_, err = db.Exec("INSERT INTO users (login,password,role) VALUES ($1,$2,$3);",login,hashPassword,role_int)
	
	if err != nil {
        fmt.Println(err)
    }

	return err
}

func ChangePasswd(db *sql.DB, login string, passwd string) error {
	var err error
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(passwd), 14) //Генерируем пароль

	if err != nil {
        fmt.Println(err)
    }

	//Записываем пароль в базу данных
	_, err = db.Exec("UPDATE users SET password = $2 WHERE login = $1;",login,hashPassword)

	if err != nil {
        fmt.Println(err)
    }
	
	return err
}

func ChangeRole(db *sql.DB, login string, role string) error {
	var role_int int
	var err error

	for i, v := range Role{ //Поиск номера роли
		if role == v {
			role_int=i
			break
		}
	}
	//Записываем роль в базу данных
	_, err = db.Exec("UPDATE users SET role = $2 WHERE login = $1;",login,role_int)

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

func DeleteUser(db *sql.DB, login string) error {
	var err error
	
	//Записываем пароль в базу данных
	_, err = db.Exec("DELETE FROM users WHERE login=$1;",login)

	if err != nil {
        fmt.Println(err)
    }

	return err
}
