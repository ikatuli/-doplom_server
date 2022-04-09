# doplom_server

По адресу `http://localhost:9000/` находится главная страница приложения.

На странице `http://localhost:9000/install` можно установить пароль администратора.

На странице `http://localhost:9000/hello`, после прохождение ацетификации можно посмотреть приветственное сообщение.

На странице `http://localhost:9000/create_user`, после прохождение ацетификации можно создать нового пользователя.

А на странице `http://localhost:9000/change_password` можно изменить пароль для текущего пользователя.

## Развёртывание

Создать сеть:

```
docker network create testnet
```

Развернуть контейнер с postgres:

```
docker run --net testnet --name postgresql -e POSTGRES_DB=test_bd -e POSTGRES_USER=test_user -e POSTGRES_PASSWORD=test_passwd -p 5432:5432 -d postgres
```

Собрать контейнер с приложением:

```
docker build --tag web_server ./
```

После сборки запускаем контейнер:

```
docker run  --net testnet  --name web_server_test  -p 9000:9000 web_server
```
