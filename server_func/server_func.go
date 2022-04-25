package server_func

import(
	"net/http" // пакет для поддержки HTTP протокола
	"html/template" // пакет для генирации html файлов
	"fmt"
	"database/sql" // пакет для работы с sql
	_ "github.com/lib/pq" // пакет драйвера PostgreSQL
	//Мои куски кода
	"doplom_server/user"
	"doplom_server/squid"
)

var db (*sql.DB) //Глобальная переменная для базы данных

func DBInit (connect string) { //Создаём в базе таблицы
	var err error
	db, err = sql.Open("postgres", connect) //Подключаемся к базе данных
	if err != nil {
        panic(err)
    }

	//fmt.Printf("%T\n", db)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id SERIAL,login character varying(20) UNIQUE,password character(60));`)

	if err != nil {
        fmt.Println("Error ", err.Error())
		panic(err)
    }
}

func DbClose () {
	db.Close()
}

func Install(w http.ResponseWriter, r *http.Request){ //Первоначальная настройка сервера
	var err error
	switch r.Method {
		case "GET":
			http.ServeFile(w, r, "./static/install.html")
		case "POST":
			if err = r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return}
			Password:= r.FormValue("passwd")
			err=user.СreateUser(db,"admin",Password)

			if err != nil {
				fmt.Fprintf(w, "Ошибка создания пользователя: %v", err)
			} else{
				fmt.Fprintf(w, "Пользователь admin создан\n")
				fmt.Fprintf(w, "Пожалуйста, отключите install mod в config.toml")
			}

		default:
			fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
		}
	}

func CreateUserUI(w http.ResponseWriter, r *http.Request){ //Создаём пользователя
	var err error
	switch r.Method {
		case "GET":
			files := []string{
						"./static/CreateUserUI.tmpl",
						"./static/base.tmpl",
					}
			t, _ := template.ParseFiles(files...)
			t.Execute(w, nil)
		case "POST":
			if err = r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return}
			Login:= r.FormValue("login")
			Password:= r.FormValue("passwd")
			err=user.СreateUser(db,Login,Password)

			if err != nil {
				fmt.Fprintf(w, "Ошибка создания пользователя: %v", err)
			} else{
				fmt.Fprintf(w, "Пользователь %s создан",Login)
			}

		default:
			fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
		}
}

func Account(w http.ResponseWriter, r *http.Request){ //Создаём пользователя
	files := []string{
						"./static/account.tmpl",
						"./static/base.tmpl",
					}
	t, _ := template.ParseFiles(files...)
	t.Execute(w, nil)
}

func ChangePassword(w http.ResponseWriter, r *http.Request){ //Меняем пароль
	login, password, err:= r.BasicAuth()

		if err {
			usr := user.FindUser(db,login)
			ok := user.CheckCredentials(usr,password)
			if ok == nil { //После того как пользователь залогинился
				var err error
				switch r.Method {
				case "GET":
					files := []string{
						"./static/ChangePassword.tmpl",
						"./static/base.tmpl",
					}
					t, _ := template.ParseFiles(files...)
					t.Execute(w, login)
				case "POST":
					if err = r.ParseForm(); err != nil {
						fmt.Fprintf(w, "ParseForm() err: %v", err)
						return}
					Password:= r.FormValue("passwd")
					err=user.ChangeUser(db,login,Password)
					
					if err != nil {
						fmt.Fprintf(w, "Ошибка при смене пароля: %v", err)
					} else{
						fmt.Fprintf(w, "Пароль пользователя %s был изменён",login)
					}
				default:
					fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
				}
				return
			}
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="Авторизация", charset="UTF-8"`)
		http.Error(w, "401 Авторизация не пройдена", http.StatusUnauthorized)
}

func Authentication(next http.HandlerFunc) http.HandlerFunc { //Логинимся перед заходом на защищённые страницы
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		login, password, err:= r.BasicAuth()

		if err {
			usr := user.FindUser(db,login)
			ok := user.CheckCredentials(usr,password)
			if ok == nil {
				next.ServeHTTP(w, r)
				return
			} 
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="Авторизация", charset="UTF-8"`)
		http.Error(w, "401 Авторизация не пройдена", http.StatusUnauthorized)
	})
}

func SquidConfig(w http.ResponseWriter, r *http.Request) {
	
	switch r.Method {
	case "GET":
		files := []string{
			"./static/squid_config.tmpl",
			"./static/base.tmpl",
		}

		t, _ := template.ParseFiles(files...)
		t.Execute(w, nil)
	case "POST":
		var err error
		if err = r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		var SquidConf = make(map[string]string)

		SquidConf["port"] = r.FormValue("port")
		SquidConf["cache"] = r.FormValue("cache")
		SquidConf["maximum_object_size"] = r.FormValue("maximum_object_size")
		if r.FormValue("SSL") !="" {
			SquidConf["SSL"] = r.FormValue("SSL")
		}

		squid.CreateConfig(SquidConf)
		/*
		*/

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
	return
}

func GenerateCertificate(w http.ResponseWriter, r *http.Request) {
	
	if r.Method == "POST" {
		var err error
		if err = r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		if r.FormValue("generate") =="Сгенерировать сертификат" {
			if err = squid.CreateCertificate(); err != nil {
				fmt.Fprintf(w, "Создание сертификата завершенно с ошибкой: %v", err)
				return
			}
		}
		fmt.Fprintf(w,"Создание сертификата завершенно")
	} else {
		fmt.Fprintf(w, "Sorry, only POST method are supported.")
	}
	return
}

func GetCertificate(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		var err error
		if err = r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		if r.FormValue("get") =="Скачать сертификат" {
			http.ServeFile(w, r, "./configuration/myCA.der")
		}
	} else {
		fmt.Fprintf(w, "Sorry, only POST method are supported.")
	}
	return
}
