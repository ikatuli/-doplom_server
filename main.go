package main

import (
    "fmt" // пакет для форматированного ввода вывода
    "net/http" // пакет для поддержки HTTP протокола
    "strings" // пакет для работы с  UTF-8 строками
    "log" // пакет для логирования
	"html/template" // пакет для генирации html файлов
	"github.com/pelletier/go-toml" // пакет для конфигурационного файла
	"database/sql" // пакет для работы с sql
	_ "github.com/lib/pq" // пакет драйвера PostgreSQL
	//Мои куски кода
	"doplom_server/user"
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
    }
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
				fmt.Fprintf(w, "Пользователь admin создан")
			}

		default:
			fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
		}
	}

func CreateUserUI(w http.ResponseWriter, r *http.Request){ //Создаём пользователя
	var err error
	switch r.Method {
		case "GET":
			http.ServeFile(w, r, "./static/CreateUserUI.html")
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

func ChangePassword(w http.ResponseWriter, r *http.Request){ //Меняем пароль
	login, password, err:= r.BasicAuth()

		if err {
			usr := user.FindUser(db,login)
			ok := user.CheckCredentials(usr,password)
			if ok == nil { //После того как пользователь залогинился
				var err error
				switch r.Method {
				case "GET":
					t, _ := template.ParseFiles("./static/ChangePassword.html")
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


func HomeRouterHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm() //анализ аргументов,
    fmt.Println(r.Form)  // ввод информации о форме на стороне сервера
    fmt.Println("path", r.URL.Path)
    fmt.Println("scheme", r.URL.Scheme)
    fmt.Println(r.Form["url_long"])
    for k, v := range r.Form {
        fmt.Println("key:", k)
        fmt.Println("val:", strings.Join(v, ""))
    }
    fmt.Fprintf(w, "Hello!") // отправляем данные на клиентскую сторону
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
	http.ServeFile(w, r, "./static/squid_config.html")
}

func main() {
	config, err := toml.LoadFile("config.toml") //Подключаем конфиг

	var host string
	var connect string

	if err != nil { //Парсим данные из конфига
    fmt.Println("Error ", err.Error())
	} else {
		host=config.Get("Server.host").(string)+":"+config.Get("Server.port").(string)
		connect="host="+config.Get("Database.dbhost").(string)+" port="+config.Get("Database.dbport").(string)+" user="+config.Get("Database.user").(string)+" password="+config.Get("Database.password").(string)+" dbname="+config.Get("Database.dbname").(string)+" sslmode=disable"
	}

	DBInit(connect) // Инициализация базы данных
	
	fileServer := http.FileServer(http.Dir("static"))
    http.Handle("/", fileServer) // установим роутер
	http.HandleFunc("/hello", Authentication(HomeRouterHandler))

	http.HandleFunc("/install",Install)// Инициализация настроек
	http.HandleFunc("/create_user",Authentication(CreateUserUI))// Создание пользователя
	http.HandleFunc("/change_password",ChangePassword)// Создание пользователя
	http.HandleFunc("/squid_config",SquidConfig)// Создание пользователя

    err = http.ListenAndServe(host, nil) // задаем слушать порт
	if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }

	defer db.Close() //Закрываем базу данных при закрытии программы
}
