package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// MySQL 连接信息（先连接到默认的 mysql 数据库）
	user := "root"
	password := "123456"
	host := "127.0.0.1"
	port := 3306

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to MySQL:", err)
	}
	defer db.Close()

	// 检查连接
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping MySQL:", err)
	}

	// 创建数据库 squads，字符集 utf8mb4，排序规则 utf8mb4_unicode_ci
	createSQL := `CREATE DATABASE IF NOT EXISTS squads CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;`
	_, err = db.Exec(createSQL)
	if err != nil {
		log.Fatal("Failed to create database:", err)
	}

	fmt.Println("✅ Database 'squads' created successfully with utf8mb4 encoding")
}
