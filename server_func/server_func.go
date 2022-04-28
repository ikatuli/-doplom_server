package server_func

import(
	"net/http" // пакет для поддержки HTTP протокола
	"html/template" // пакет для генирации html файлов
	"fmt"
	"os" //Для проверки существования файлов
	"errors" 
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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id SERIAL,login character varying(20) UNIQUE,password character(60),role smallint);`)

	if err != nil {
        fmt.Println("Error ", err.Error())
		panic(err)
    }

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS squid (id SERIAL,parameter character varying(25) UNIQUE,value character varying(25));`)

	if err != nil {
        fmt.Println("Error ", err.Error())
		panic(err)
    }
}

func DbClose () {
	db.Close()
}

func report (w http.ResponseWriter,s string) { //Ввывод всяческих сообщений о статусе
	files := []string{
		"./static/report.tmpl",
		"./static/base.tmpl",
	}
	t, _ := template.ParseFiles(files...)
	t.Execute(w, s)
}

func Install(w http.ResponseWriter, r *http.Request){ //Первоначальная настройка сервера
	switch r.Method {
		case "GET":
			http.ServeFile(w, r, "./static/install.html")
		case "POST":
			var err error
			if err = r.ParseForm(); err != nil {
			report(w,fmt.Sprintf("ParseForm() err: %v", err))
			return}
			Password:= r.FormValue("passwd")
			err=user.СreateUser(db,"admin",Password,"admin")			

			if err != nil {
				report(w,fmt.Sprintf("Ошибка создания пользователя: %v", err))
			} else{
				report(w,"Пользователь admin создан. Пожалуйста, отключите install mod в config.toml")
			}

		default:
			report(w,"Sorry, only GET and POST methods are supported.")
		}
	}

func CreateUserUI(w http.ResponseWriter, r *http.Request, userProfile  user.User){ //Создаём пользователя
	if userProfile.Role != "admin" {report(w,"Нет прав на доступ к этой странице");return}
	switch r.Method {
		case "GET":
			files := []string{
						"./static/CreateUserUI.tmpl",
						"./static/base.tmpl",
					}
			t, _ := template.ParseFiles(files...)
			t.Execute(w, user.Role)
		case "POST":
			var err error
			if err = r.ParseForm(); err != nil {
			report(w,fmt.Sprintf("ParseForm() err: %v", err))
			return}
			err=user.СreateUser(db,r.FormValue("login"),r.FormValue("passwd"),r.FormValue("role"))

			if err != nil {
				report(w,fmt.Sprintf("Ошибка создания пользователя: %v", err))
			} else{
				report(w,fmt.Sprintf( "Пользователь %s создан",r.FormValue("login")))
			}

		default:
			report(w,"Sorry, only GET and POST methods are supported.")
		}
}

func Account(w http.ResponseWriter, r *http.Request, userProfile  user.User){ //Создаём пользователя
	files := []string{
						"./static/account.tmpl",
						"./static/base.tmpl",
					}
	t, _ := template.ParseFiles(files...)
	t.Execute(w, userProfile.Role)
}

func ChangePassword(w http.ResponseWriter, r *http.Request, userProfile  user.User){ //Меняем пароль
	switch r.Method {
	case "GET":
		files := []string{
			"./static/ChangePassword.tmpl",
			"./static/base.tmpl",
		}
		t, _ := template.ParseFiles(files...)
		t.Execute(w, userProfile.Login)
	case "POST":
		var err error
		if err = r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return}
		Password:= r.FormValue("passwd")
		err=user.ChangeUser(db,userProfile.Login,Password)
					
		if err != nil {
			report(w, fmt.Sprintf( "Ошибка при смене пароля: %v", err))
		} else{
			report(w,fmt.Sprintf("Пароль пользователя %s был изменён",userProfile.Login))
		}
	default:
		report(w,"Sorry, only GET and POST methods are supported.")
	}
	return
}

func Authentication(next func(http.ResponseWriter,*http.Request,user.User)) http.HandlerFunc { //Логинимся перед заходом на защищённые страницы
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		login, password, err:= r.BasicAuth()

		if err {
			usr := user.FindUser(db,login)
			ok := user.CheckCredentials(usr,password)
			if ok == nil {
				next(w, r, usr)
				return
			} 
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="Авторизация", charset="UTF-8"`)
		http.Error(w, "401 Авторизация не пройдена", http.StatusUnauthorized)
	})
}

func SquidConfig(w http.ResponseWriter, r *http.Request, userProfile  user.User) {
	if userProfile.Role != "admin" {report(w,"Нет прав на доступ к этой странице");return}
	switch r.Method {
	case "GET":

		rows, err:=db.Query("SELECT * FROM squid;")
		if err != nil {	fmt.Println(err)}

		//Инициализируем переменные
		var Conf squid.SquidConf
		Conf.Port= "";Conf.Cache = "";Conf.MaximumObjectSize = "";Conf.SSL = ""

		//Вытаскиваем из запроса
		var id string;var parameter string;var value string;
		for rows.Next() {
			if err = rows.Scan(&id,&parameter,&value); err != nil {
				fmt.Println(err)
			}
			switch parameter {
			case "Port":
				Conf.Port=value
			case "Cache":
				Conf.Cache=value
			case "MaximumObjectSize":
				Conf.MaximumObjectSize=value
			case "SSL":
				Conf.SSL=value
			}
		}

		if Conf.Port=="" {Conf.Port= "3128"}
		if Conf.Cache=="" {Conf.Cache = "64"}
		if Conf.MaximumObjectSize=="" {Conf.MaximumObjectSize= "10"}

		files := []string{
			"./static/squid_config.tmpl",
			"./static/base.tmpl",
		}

		t, _ := template.ParseFiles(files...)
		t.Execute(w, Conf)
	case "POST":
		var err error
		if err = r.ParseForm(); err != nil {
			report(w,fmt.Sprintf("ParseForm() err: %v", err))
			return
		}

		var Conf squid.SquidConf
		Conf.Port = r.FormValue("port")
		Conf.Cache = r.FormValue("cache")
		Conf.MaximumObjectSize = r.FormValue("maximum_object_size")
		if r.FormValue("SSL") !="" {
			Conf.SSL = r.FormValue("SSL")
		}

		if Conf.SSL == "SSL" { //Сгенерировать сертификат если его нет
			if _, err := os.Stat("/etc/squid/myCA.pem"); errors.Is(err, os.ErrNotExist) {
				squid.CreateCertificate()
			}
		}
		_, err = db.Exec("INSERT INTO squid (parameter,value) VALUES ($1,$2) ON CONFLICT (parameter) DO UPDATE SET value = $2;","Port",Conf.Port)
		if err != nil {fmt.Println(err)}

		_, err = db.Exec("INSERT INTO squid (parameter,value) VALUES ($1,$2) ON CONFLICT (parameter) DO UPDATE SET value = $2;","Cache",Conf.Cache)
		if err != nil {fmt.Println(err)}

		_, err = db.Exec("INSERT INTO squid (parameter,value) VALUES ($1,$2) ON CONFLICT (parameter) DO UPDATE SET value = $2;","MaximumObjectSize",Conf.MaximumObjectSize)
		if err != nil {fmt.Println(err)}

		_, err = db.Exec("INSERT INTO squid (parameter,value) VALUES ($1,$2) ON CONFLICT (parameter) DO UPDATE SET value = $2;","SSL",Conf.SSL)
		if err != nil {fmt.Println(err)}

		squid.CreateConfig(Conf)

	default:
		report(w,"Sorry, only GET and POST methods are supported.")
	}
	return
}

func GenerateCertificate(w http.ResponseWriter, r *http.Request, userProfile  user.User) {
	if userProfile.Role != "admin" {report(w,"Нет прав на доступ к этой странице");return}
	if r.Method == "POST" {
		var err error
		if err = r.ParseForm(); err != nil {
			report(w,fmt.Sprintf("ParseForm() err: %v", err))
			return
		}

		if r.FormValue("generate") =="Сгенерировать сертификат" {
			if err = squid.CreateCertificate(); err != nil {
				report(w,fmt.Sprintf("Создание сертификата завершенно с ошибкой: %v", err))
				return
			}
		}
		report(w,"Создание сертификата завершенно")
	} else {
		report(w,"Sorry, only POST method are supported.")
	}
	return
}

func GetCertificate(w http.ResponseWriter, r *http.Request, userProfile  user.User) {
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
		report(w,"Sorry, only POST method are supported.")
	}
	return
}
