package rule

import(
	"fmt"
	"os"
	"os/exec"
	"database/sql" // пакет для работы с sql
	_ "github.com/lib/pq" // пакет драйвера PostgreSQL
	"doplom_server/squid"
)


func FindRuleID(db *sql.DB,name string) int{
	//Запрашиваем ip профиля
	rows, err:=db.Query("SELECT id FROM rule WHERE name = $1;",name)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var id int
	rows.Next()
	if err = rows.Scan(&id); err != nil {
			fmt.Println(err)
		}

	return id
}

func CreateRuleID(db *sql.DB, name string, creator string) error {
	var err error

	//Записываем название профиля в базу данных
	_, err = db.Exec("INSERT INTO rule (name,creator) VALUES ($1,$2);",name,creator)
	
	if err != nil {
        fmt.Println(err)
    }

	return err
}

func ChangeRuleName (db *sql.DB,id int, name string) error {
	var err error

	//Записываем пароль в базу данных
	_, err = db.Exec("UPDATE rule SET name = $2 WHERE id = $1;",id,name)

	if err != nil {
        fmt.Println(err)
    }

	return err
}


func ChangeDomain(db *sql.DB, id int, domain string) error {
	var err error
	//Записываем домен в базу данных
	_, err = db.Exec("INSERT INTO rule_domain (rule_id,domain) VALUES ($1,$2) ON CONFLICT (rule_id) DO UPDATE SET domain = $2;",id,domain)

	if err != nil {
        fmt.Println(err)
    }

	return err
}

func FindDomain(db *sql.DB,id int) string{
	//Запрашиваем домен
	rows, err:=db.Query("SELECT domain FROM  rule_domain WHERE rule_id = $1;",id)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var domain string
	rows.Next()
	if err = rows.Scan(&domain); err != nil {
			fmt.Println(err)
		}

	return domain
}

func DeleteRule (db *sql.DB,name string) error {
	var err error

	var id = FindRuleID(db,name)
	//Удаляем все данные из связанных таблиц.
	_, err = db.Exec("DELETE FROM rule_domain WHERE rule_id = $1;",id)

	if err != nil {
        fmt.Println(err)
    }

	//Удаляем правило
	_, err = db.Exec("DELETE FROM rule WHERE id = $1;",id)

	if err != nil {
        fmt.Println(err)
    }

	return err
}

func Activate(db *sql.DB, name string) error {
	var err error
	var id = FindRuleID(db,name)

	f, err := os.Create("./configuration/deniedsites.squid")

	if err != nil {
		return err
	}
	defer f.Close()
	
	f.WriteString(FindDomain(db,id))

	err = exec.Command("mv","./configuration/deniedsites.squid", "/etc/squid/deniedsites.squid").Run()
    if err != nil {
        return err
    }

	squid.DeliteCache()

	if (squid.Status() == "active") {
		squid.Start("restart")
	}

	return err
}
