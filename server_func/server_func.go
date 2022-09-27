package server_func

import(
	"net/http" // пакет для поддержки HTTP протокола
	"html/template" // пакет для генирации html файлов
	"fmt"
	"os" //Для проверки существования файлов
	"errors" 
	"database/sql" // пакет для работы с sql
	_ "github.com/lib/pq" // пакет драйвера PostgreSQL
	"strconv" //Преобразование чисел в строки и наоборот
	//Мои куски кода
	"doplom_server/user"
	"doplom_server/squid"
	"doplom_server/e2guardian"
	"doplom_server/clamav"
	"doplom_server/dnscrypt"
	"doplom_server/rule"
)

var db (*sql.DB) //Глобальная переменная для базы данных

func DBInit (connect string) { //Создаём в базе таблицы
	var err error
	db, err = sql.Open("postgres", connect) //Подключаемся к базе данных
	if err != nil {
        panic(err)
    }

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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS etoguardian (id SERIAL,parameter character varying(25) UNIQUE,value character varying(25));`)
	
	if err != nil {
        fmt.Println("Error ", err.Error())
	panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS dnscrypt (id SERIAL,parameter character varying(25) UNIQUE,value character varying(25));`)

	if err != nil {
        fmt.Println("Error ", err.Error())
		panic(err)
    }

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS rule (id SERIAL,creator character varying(20),name character varying(20) UNIQUE);`)

	if err != nil {
        fmt.Println("Error ", err.Error())
		panic(err)
    }
	
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS rule_domain (id SERIAL,rule_id integer UNIQUE, domain text);`)

	if err != nil {
        fmt.Println("Error ", err.Error())
		panic(err)
    }
}

func DbClose () {
	db.Close()
}

func Index (w http.ResponseWriter, r *http.Request, userProfile  user.User) {
	files := []string{
		"./static/index.tmpl",
		"./static/base.tmpl",
	}
	t, _ := template.ParseFiles(files...)

	var status = map[string]string {
		"squid": squid.Status(),
		"e2guardian": e2guardian.Status(),
		"clamav": clamav.Status(),
		"dnscrypt": dnscrypt.Status(),
	}

	t.Execute(w, status)
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
			var test string
			if err = r.ParseForm(); err != nil {
			report(w,fmt.Sprintf("ParseForm() err: %v", err))
			return}
			Password:= r.FormValue("passwd")
			err=user.СreateUser(db,"admin",Password,"admin")			

			if err != nil {
				test=fmt.Sprintf("Ошибка создания пользователя: %v\n", err)
			} else{
				test="Пользователь admin создан.\n"
			}

			out,err:=clamav.Update()
			
			if err != nil {
				test=test+fmt.Sprintf("База данных антивируса не обновлена: %v\n", out)
			} else{
				test=test+"База данных антивируса обновлена\n"
			}
			
			report(w,test+"Пожалуйста, отключите install mod в config.toml")
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

func ChangeUsers(w http.ResponseWriter, r *http.Request, userProfile  user.User){ //Создаём пользователя
	if userProfile.Role != "admin" {report(w,"Нет прав на доступ к этой странице");return}
	switch r.Method {
		case "GET":
			rows, err:=db.Query("SELECT login FROM users;")
			if err != nil {	fmt.Println(err)}
			defer rows.Close()
			//Инициализируем список пользователь, сразу создавая 2 ячейки.
			users :=make([]string, 0, 2)
			//Вытаскиваем из запроса
			var login string;
			for rows.Next() {
				if err = rows.Scan(&login); err != nil {
					fmt.Println(err)
				}
				users=append(users,login)
			}

			files := []string{
						"./static/ChangeUsers.tmpl",
						"./static/base.tmpl",
					}
			t, _ := template.ParseFiles(files...)
			t.Execute(w, users)
		case "POST":
			var err error
			if err = r.ParseForm(); err != nil {
			report(w,fmt.Sprintf("ParseForm() err: %v", err))
			return}
			switch r.FormValue("action"){
				case "delete":
					err = user.DeleteUser(db, r.FormValue("login"))
					if err != nil {
						report(w,fmt.Sprintf("Ошибка удаления пользователя: %v", err))
					} else{
						report(w,fmt.Sprintf( "Пользователь %s удалён",r.FormValue("login")))
					}
				case "password":
					http.Redirect(w, r, "/change_password?login="+r.FormValue("login"), http.StatusFound)
				case "role":
					http.Redirect(w, r, "/change_role?login="+r.FormValue("login"), http.StatusFound)
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

		if (r.URL.RawQuery !="") && (userProfile.Role == "admin") {
			userProfile.Login=r.URL.Query().Get("login")
		}

		t.Execute(w, userProfile.Login)
	case "POST":
		var err error
		if err = r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return}
		//Если запрос послал администратор, то можно подменить логин
		//А иначе будет изменён пароль отправившего запрос пользователя.
		if userProfile.Role == "admin" {
			userProfile.Login=r.FormValue("login")
		}

		Password:= r.FormValue("passwd")
		err=user.ChangePasswd(db,userProfile.Login,Password)
					
		if err != nil {
			report(w, fmt.Sprintf( "Ошибка при изменении пароля: %v", err))
		} else{
			report(w,fmt.Sprintf("Пароль пользователя %s был изменении",userProfile.Login))
		}
	default:
		report(w,"Sorry, only GET and POST methods are supported.")
	}
	return
}

func ChangeRole(w http.ResponseWriter, r *http.Request, userProfile  user.User){ //Меняем пароль
	if userProfile.Role != "admin" {report(w,"Нет прав на доступ к этой странице");return}
	switch r.Method {
	case "GET":
		files := []string{
			"./static/ChangeRole.tmpl",
			"./static/base.tmpl",
		}
		t, _ := template.ParseFiles(files...)

		login:=r.URL.Query().Get("login")
		Data := struct {
			Login string
			Role string
			RoleList []string
		}{
			Login: login,
			Role: user.FindUser(db,login).Role,
			RoleList: user.Role,
		}
		t.Execute(w, Data)
	case "POST":
		var err error
		if err = r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return}
		
		err=user.ChangePasswd(db,r.FormValue("login"),r.FormValue("role"))
					
		if err != nil {
			report(w, fmt.Sprintf( "Ошибка при изменении роли: %v", err))
		} else{
			report(w,fmt.Sprintf("Роль пользователя %s была изменениа",r.FormValue("login")))
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
	//Если пользователь не администратор, то ему не показываются все настройки.
	if userProfile.Role != "admin" {
		var conf = map[string]string {
			"role": userProfile.Role,
		}
		
		files := []string{
			"./static/squid_config.tmpl",
			"./static/base.tmpl",
		}

		t, _ := template.ParseFiles(files...)
		t.Execute(w, conf)
		return
	}

	switch r.Method {
	case "GET":
		rows, err:=db.Query("SELECT * FROM squid;")
		if err != nil {	fmt.Println(err)}
		defer rows.Close()

		//ИнGициализируем переменные с данными о настройках
		var conf = make(map[string]string)

		//Вытаскиваем из запроса
		var id string;var parameter string;var value string;
		for rows.Next() {
			if err = rows.Scan(&id,&parameter,&value); err != nil {
				fmt.Println(err)
			}
			conf[parameter] = value
		}

		//Вносим значение по умолчанию
		if _, err:= conf["Port"]; !err { conf ["Port"] = "3128"	}
		if _, err:= conf["Cache"]; !err { conf ["Cache"] = "64"	}
		if _, err:= conf["MaximumObjectSize"]; !err { conf ["MaximumObjectSize"] = "10"	}
		if _, err:= conf["SSL"]; !err { conf ["SSL"] = ""	}
		if _, err:= conf["e2guardian"]; !err { conf ["e2guardian"] = ""	}
		if _, err:= conf["DNS"]; !err { conf ["DNS"] = ""	}

		//Роль пользователя
		conf ["role"] = userProfile.Role
		conf["on"] = squid.Status()

		files := []string{
			"./static/squid_config.tmpl",
			"./static/base.tmpl",
		}

		t, _ := template.ParseFiles(files...)
		t.Execute(w, conf)
	case "POST":
		var err error
		if err = r.ParseForm(); err != nil {
			report(w,fmt.Sprintf("ParseForm() err: %v", err))
			return
		}

		//Инициализируем переменные с данными о настройках
		var conf = make(map[string]string)

		//Вытаскиваем из формы все параметры
		for parameter, value:=range r.Form {
			conf[parameter]=value[0]
		}

		if conf["SSL"] == "SSL" { //Сгенерировать сертификат если его нет
			if _, err := os.Stat("/etc/squid/myCA.pem"); errors.Is(err, os.ErrNotExist) {
				squid.CreateCertificate()
			}
		} else {
			conf["SSL"] = "" //Обнуляем значение
		}

		if _, err:= conf["e2guardian"]; !err { conf ["e2guardian"] = ""}
		if _, err:= conf["DNS"]; !err { conf ["DNS"] = ""	}

		//Сохраняем в базу данных все параметры
		for parameter, value:=range conf {
			_, err = db.Exec("INSERT INTO squid (parameter,value) VALUES ($1,$2) ON CONFLICT (parameter) DO UPDATE SET value = $2;",parameter,value)
			if err != nil {fmt.Println(err)}
		}
		
		err=squid.CreateConfig(conf)

		if err != nil {
			report(w,fmt.Sprintf("Ошибка запуска прокси сервера: %v", err))
			return
		} else {
			report(w,"Настройки прокси сервера обновлены")
		}


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

func Journal (w http.ResponseWriter, r *http.Request, userProfile  user.User) {
	if userProfile.Role != "admin" {report(w,"Нет прав на доступ к этой странице");return}
	if r.Method == "GET" {
		files := []string{
		"./static/journal.tmpl",
		"./static/base.tmpl",
	}
	t, _ := template.ParseFiles(files...)

	var status = map[string]string {
		"squid": squid.Journal(),
		"e2guardian": e2guardian.Journal(),
		"clamav": clamav.Journal(),
	}

	t.Execute(w, status)
	} else {
		report(w,"Sorry, only GET method are supported.")
	}
	return
}

func E2guardianConfig(w http.ResponseWriter, r *http.Request, userProfile  user.User) {
	if userProfile.Role != "admin" {report(w,"Нет прав на доступ к этой странице");return}
	switch r.Method {
	case "GET":
		rows, err:=db.Query("SELECT * FROM etoguardian;")
		if err != nil {	fmt.Println(err)}
		defer rows.Close()

		//Инициализируем переменные с данными о настройках
		var conf = make(map[string]string)

		//Вытаскиваем из запроса
		var id string;var parameter string;var value string;
		for rows.Next() {
			if err = rows.Scan(&id,&parameter,&value); err != nil {
				fmt.Println(err)
			}
			conf[parameter] = value
		}
		
		conf["on"] = e2guardian.Status()

		//Вносим значение по умолчанию
		if _, err:= conf["ClamAV"]; !err { conf ["ClamAV"] = ""}
		
		files := []string{
			"./static/e2guardian_config.tmpl",
			"./static/base.tmpl",
		}

		t, _ := template.ParseFiles(files...)
		t.Execute(w, conf)
	case "POST":
		var err error
		if err = r.ParseForm(); err != nil {
			report(w,fmt.Sprintf("ParseForm() err: %v", err))
			return
		}

		//Инициализируем переменные с данными о настройках
		var conf = make(map[string]string)

		//Вытаскиваем из формы все параметры
		for parameter, value:=range r.Form {
			conf[parameter]=value[0]
		}

		if _, err:= conf["ClamAV"]; !err { conf ["ClamAV"] = ""}

		//Сохраняем в базу данных все параметры
		for parameter, value:=range conf {
			_, err = db.Exec("INSERT INTO etoguardian (parameter,value) VALUES ($1,$2) ON CONFLICT (parameter) DO UPDATE SET value = $2;",parameter,value)
			if err != nil {fmt.Println(err)}
		}

		err=e2guardian.CreateConfig(conf)

		if err != nil {
			report(w,fmt.Sprintf("Ошибка запуска прокси сервера: %v", err))
			return
		} else {
			report(w,"Настройки прокси сервера обновлены")
		}


	default:
		report(w,"Sorry, only GET and POST methods are supported.")
	}
	return
}

func ClamavConfig(w http.ResponseWriter, r *http.Request, userProfile  user.User) {
	if userProfile.Role != "admin" {report(w,"Нет прав на доступ к этой странице");return}
	switch r.Method {
	case "GET":

		var conf = map[string]string {
		"on": clamav.Status(),
		}

		files := []string{
			"./static/clamav_config.tmpl",
			"./static/base.tmpl",
		}

		t, _ := template.ParseFiles(files...)
		t.Execute(w, conf)
	case "POST":
		var err error
		if err = r.ParseForm(); err != nil {
			report(w,fmt.Sprintf("ParseForm() err: %v", err))
			return
		}

		out,err:=clamav.Update()
			
		if err != nil {
			report(w,fmt.Sprintf("База данных антивируса не обновлена: %v\n", out))
		} else{
			report(w,"База данных антивируса обновлена\n")
		}

	default:
		report(w,"Sorry, only GET and POST methods are supported.")
	}
	return
}


func Service(w http.ResponseWriter, r *http.Request, userProfile  user.User) {
	if userProfile.Role != "admin" {report(w,"Нет прав на доступ к этой странице");return}
	if r.Method == "POST" {
		var err error
		if err = r.ParseForm(); err != nil {
			report(w,fmt.Sprintf("ParseForm() err: %v", err))
			return
		}

		var err1 error
		var err2 error
		
		switch r.FormValue("service") {
			case "e2guardian":
				if  r.FormValue("on") == "on"{
					err1=e2guardian.Start("enable")
					err2=e2guardian.Start("start")
				} else {
					err1=e2guardian.Start("stop")
					err2=e2guardian.Start("disable")
				}
			case "squid":
				if  r.FormValue("on") == "on"{
					err1=squid.Start("enable")
					err2=squid.Start("start")
				} else {
					err1=squid.Start("stop")
					err2=squid.Start("disable")
				}
			case "clamav":
				if  r.FormValue("on") == "on"{
					err1=clamav.Start("enable")
					err2=clamav.Start("start")
				} else {
					err1=clamav.Start("stop")
					err2=clamav.Start("disable")
				}
			case "dnscrypt":
				if  r.FormValue("on") == "on"{
					err1=dnscrypt.Start("enable")
					err2=dnscrypt.Start("start")
				} else {
					err1=dnscrypt.Start("stop")
					err2=dnscrypt.Start("disable")
				}
		}

		if err1 != nil || err2 != nil {
			report(w,fmt.Sprintf("Действие не выполнено: %v , %v", err1, err2))
			return
		}
		
		report(w,"Действие выполнено")
	} else {
		report(w,"Sorry, only POST method are supported.")
	}

	return
}

func DnscryptConfig (w http.ResponseWriter, r *http.Request, userProfile  user.User) {
	if userProfile.Role != "admin" {report(w,"Нет прав на доступ к этой странице");return}
		switch r.Method {
	case "GET":
		rows, err:=db.Query("SELECT * FROM dnscrypt;")
		if err != nil {	fmt.Println(err)}
		defer rows.Close()

		//Инициализируем переменные с данными о настройках
		var conf = make(map[string]string)

		//Вытаскиваем из запроса
		var id string;var parameter string;var value string;
		for rows.Next() {
			if err = rows.Scan(&id,&parameter,&value); err != nil {
				fmt.Println(err)
			}
			conf[parameter] = value
		}
		
		conf["on"] = dnscrypt.Status()

		//Вносим значение по умолчанию
		if _, err:= conf["ClamAV"]; !err { conf ["ClamAV"] = ""}
		if _, err:= conf["IPv6"]; !err { conf ["IPv6"] = ""}
		if _, err:= conf["Cache"]; !err { conf ["Cache"] = "4096"}
		if _, err:= conf["Timeout"]; !err { conf ["Timeout"] = "5000"}
		
		files := []string{
			"./static/dnscrypt_config.tmpl",
			"./static/base.tmpl",
		}

		t, _ := template.ParseFiles(files...)
		t.Execute(w, conf)

	case "POST":
		var err error
		if err = r.ParseForm(); err != nil {
			report(w,fmt.Sprintf("ParseForm() err: %v", err))
			return
		}

		//Инициализируем переменные с данными о настройках
		var conf = make(map[string]string)

		//Вытаскиваем из формы все параметры
		for parameter, value:=range r.Form {
			conf[parameter]=value[0]
		}

		if _, err:= conf["ClamAV"]; !err { conf ["ClamAV"] = ""}
		if _, err:= conf["IPv6"]; !err { conf ["IPv6"] = ""}

		//Сохраняем в базу данных все параметры
		for parameter, value:=range conf {
			_, err = db.Exec("INSERT INTO dnscrypt (parameter,value) VALUES ($1,$2) ON CONFLICT (parameter) DO UPDATE SET value = $2;",parameter,value)
			if err != nil {fmt.Println(err)}
		}

		err=dnscrypt.CreateConfig(conf)

		if err != nil {
			report(w,fmt.Sprintf("Ошибка запуска прокси сервера: %v", err))
			return
		} else {
			report(w,"Настройки прокси сервера обновлены")
		}		

	default:
		report(w,"Sorry, only GET and POST methods are supported.")
	}
	return
}

func RuleMain(w http.ResponseWriter, r *http.Request, userProfile  user.User){ //Выбрать правило
	switch r.Method {
	case "GET":
		rows, err:=db.Query("SELECT name FROM rule;")
		if err != nil {	fmt.Println(err)}
		defer rows.Close()

		//Инициализируем переменные с данными о настройках
		var rule_list = make([]string,0)

		//Вытаскиваем имя из запроса
		var name string;
		for rows.Next() {
			if err = rows.Scan(&name); err != nil {
				fmt.Println(err)
			}
			rule_list= append(rule_list,name)
		}

		
		files := []string{
			"./static/rule_main.tmpl",
			"./static/base.tmpl",
		}

		t, _ := template.ParseFiles(files...)
		t.Execute(w, rule_list)

	case "POST":
		var err error
		if err = r.ParseForm(); err != nil {
			report(w,fmt.Sprintf("ParseForm() err: %v", err))
			return
		}

		switch r.Form["action"][0] {
			case "activate":
				err= rule.Activate(db,r.Form["rule"][0])
				if err != nil {
					report(w,fmt.Sprintf("Активация правила завершено с ошибкой: %v", err))
					return
				} else {
					report(w,"Правило активировано")
				}
			case "delete":
				err= rule.DeleteRule(db,r.Form["rule"][0])
				if err != nil {
					report(w,fmt.Sprintf("Удаление правила завершено с ошибкой: %v", err))
					return
				} else {
					report(w,"Правило удаленно")
				}
			}


	default:
		report(w,"Sorry, only GET and POST methods are supported.")
	}
	return
}

func RuleCreate(w http.ResponseWriter, r *http.Request, userProfile  user.User){ //Выбрать правило
	switch r.Method {
	case "GET":

		var conf = make(map[string]string)
	
		conf["Name"]= ""

		if (r.URL.RawQuery !=""){
			conf["Name"]=r.URL.Query().Get("name")
		}

		if (conf["Name"] !="" ){
			var id =rule.FindRuleID(db,conf["Name"])
			conf["Id"]=strconv.Itoa(id)
			conf["Domain"]=rule.FindDomain(db,id)
		} else {
			conf["Domain"]=""
			conf["Id"]= "0"
		}

		files := []string{
			"./static/rule_create.tmpl",
			"./static/base.tmpl",
		}

		t, _ := template.ParseFiles(files...)
		t.Execute(w, conf)

	case "POST":
		var err error
		if err = r.ParseForm(); err != nil {
			report(w,fmt.Sprintf("ParseForm() err: %v", err))
			return
		}

		//Инициализируем переменные с данными о настройках
		var conf = make(map[string]string)

		//Вытаскиваем из формы все параметры
		for parameter, value:=range r.Form {
			conf[parameter]=value[0]
		}

		var id int;
		id,err = strconv.Atoi(conf["id"])
		
		if (id==0) { //Создать id если его нет
			if err = rule.CreateRuleID(db,conf["name"],userProfile.Login); err != nil {
				report(w,fmt.Sprintf("Создание правила завершенно с ошибкой: %v", err))
				return
			} else {
				id = rule.FindRuleID(db,conf["name"])
			}
		} 

		rule.ChangeRuleName(db,id,conf["name"])
		if err != nil {fmt.Println(err)}

		rule.ChangeDomain(db,id,conf["domain"])
		if err != nil {fmt.Println(err)}

		if err != nil {
			report(w,fmt.Sprintf("Ошибка создания правила: %v", err))
			return
		} else {
			report(w,"Правило создано")
		}		

	default:
		report(w,"Sorry, only GET and POST methods are supported.")
	}
	return
}
