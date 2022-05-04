package main

import (
    "fmt" // пакет для форматированного ввода вывода
    "net/http" // пакет для поддержки HTTP протокола
    "log" // пакет для логирования
	"github.com/pelletier/go-toml" // пакет для конфигурационного файла
	"doplom_server/server_func"
)

func finished() {
    fmt.Println("Finished finding largest")
}

func main() {
	config, err := toml.LoadFile("config.toml") //Подключаем конфиг

	var host string
	var connect string
	var installmod bool

	if err != nil { //Парсим данные из конфига
    fmt.Println("Error ", err.Error())
	} else {
		host=config.Get("Server.host").(string)+":"+config.Get("Server.port").(string)
		connect="host="+config.Get("Database.dbhost").(string)+" port="+config.Get("Database.dbport").(string)+" user="+config.Get("Database.user").(string)+" password="+config.Get("Database.password").(string)+" dbname="+config.Get("Database.dbname").(string)+" sslmode=disable"
		installmod=config.Get("Server.installmod").(bool)
	}

	server_func.DBInit(connect) // Инициализация базы данных

	http.HandleFunc("/",server_func.Authentication(server_func.Index))

	if installmod {
		http.HandleFunc("/install",server_func.Install)// Инициализация настроек
	}
	
	http.HandleFunc("/account",server_func.Authentication(server_func.Account)) //Аккаунт
		http.HandleFunc("/create_user",server_func.Authentication(server_func.CreateUserUI))// Создание пользователя
		http.HandleFunc("/change_password",server_func.Authentication(server_func.ChangePassword))// Создание пользователяa

	http.HandleFunc("/squid_config",server_func.Authentication(server_func.SquidConfig))// Настройки прокси сервера
		http.HandleFunc("/generate_certificate",server_func.Authentication(server_func.GenerateCertificate))// Создать сертификат
		http.HandleFunc("/get_certificate",server_func.Authentication(server_func.GetCertificate))// Создать сертификат

	http.HandleFunc("/journal",server_func.Authentication(server_func.Journal))// Настройки прокси сервера

    err = http.ListenAndServe(host, nil) // Указываем адресс и порт
	if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }

	defer server_func.DbClose() //Закрываем базу данных при закрытии программы
}
