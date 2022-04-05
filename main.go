package main

import (
    "fmt" // пакет для форматированного ввода вывода
    "net/http" // пакет для поддержки HTTP протокола
    "strings" // пакет для работы с  UTF-8 строками
    "log" // пакет для логирования
	"github.com/pelletier/go-toml" // пакет для конфигурационного файла
	//Мои куски кода
	"doplom_server/user"
)

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

func Authentication(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		login, password, err:= r.BasicAuth()

		if err {
			usr := user.FindUser(login)
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

func main() {
	config, err := toml.LoadFile("config.toml")

	var host string

	if err != nil {
    fmt.Println("Error ", err.Error())
	} else {
		host=config.Get("Server.host").(string)+":"+config.Get("Server.port").(string)
	}

	fileServer := http.FileServer(http.Dir("static"))
    http.Handle("/", fileServer) // установим роутер
	http.HandleFunc("/hello", Authentication(HomeRouterHandler))

    err = http.ListenAndServe(host, nil) // задаем слушать порт
	if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
