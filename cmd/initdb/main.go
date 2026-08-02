package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:sta_go@tcp(121.40.235.227:3307)/?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	for _, name := range []string{"IM_Toy"} {
		_, err := db.Exec("CREATE DATABASE IF NOT EXISTS `" + name + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci")
		if err != nil {
			log.Fatalf("create %s: %v", name, err)
		}
		fmt.Printf("database %s ensured\n", name)
	}
}
