package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"sync"
)

var User = flag.String("username", "root", "Database user")
var Password = flag.String("password", "", "Database password")
var Host = flag.String("host", "127.0.0.1", "Database host")
var Port = flag.Int("port", 3306, "Database port")
var Database = flag.String("database", "bench_db", "Database name")
var Tables = flag.Int("tables", 10, "Number of tables to generate")
var Records = flag.Int("records", 100, "Number of records to generate")
var Threads = flag.Int("threads", 1, "Number of threads used for generation")

func main() {
	flag.Parse()

	log.SetFlags(log.Lshortfile | log.Ltime)

	log.Println("Running Database generator tool")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema", *User, *Password, *Host, *Port))

	if err != nil {
		log.Fatalln("Failed to open database connection", err)
	}

	defer db.Close()

	//create database if not exists

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + *Database) //todo prepare? cant prolly
	log.Println("CREATE DATABASE IF NOT EXISTS " + *Database)

	if err != nil {
		log.Fatalln("Failed to prepare", err)
	}

	_, err = db.Exec("use " + *Database)

	if err != nil {
		log.Fatalln(err)
	}

	//generate tables

	data := make(chan int, 1)

	wg := &sync.WaitGroup{}

	for i := 0; i < *Threads; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, data chan int) {
			defer wg.Done()

			c := getConn()

			for i := range data {
				generate(c, i, *Records)
			}

		}(wg, data)
	}

	for i := 0; i < *Tables; i++ {
		data <- i
	}
	close(data)
	wg.Wait()

	log.Println("Finished")
	//prepare workers

	//prepare input channel

}

func getConn() *sql.DB {

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema", *User, *Password, *Host, *Port))

	if err != nil {
		log.Fatalln("Failed to open database connection", err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatalln("Failed to ping new connection", err)
	}

	_, err = db.Exec("use " + *Database)

	if err != nil {
		log.Fatalln("Connection opening error:", err)
	}

	return db
}

func generate(db *sql.DB, id int, rows int) {

	//create table

	tableName := "test_" + strconv.Itoa(id)

	createTable(db, tableName)

	//generate data

	fillTable(db, rows, tableName)

}

func fillTable(db *sql.DB, rows int, tableName string) {
	sqlStmt := "INSERT INTO `" + tableName + "` (`id`, `application`, `users`, `is_active`) VALUES (NULL, 'test_app_1', '4343453', '1'), (NULL, 'test_app_2', '776556756', '1'), (NULL, 'test_app_3', '8888674', '1'), (NULL, 'test_app_4', '34588', '1'), (NULL, 'test_app_5', '888', '1'), (NULL, 'test_app_6', '87979745', '1'), (NULL, 'test_app_7', '0', '1'), (NULL, 'test_app_8', '663334345', '1'), (NULL, 'test_app_9', '9', '1'), (NULL, 'test_app_10c', '-22', '1')"

	for i := 0; i < rows; i += 10 {
		//insert 10 rows
		_, err := db.Exec(sqlStmt)
		if err != nil {
			log.Fatalln("Failed inserting into table", tableName, err)
		}
	}
}

func createTable(db *sql.DB, name string) {

	sqlStmt := "CREATE TABLE IF NOT EXISTS `" + *Database + "`.`" + name + "` ( `id` INT(11) NOT NULL AUTO_INCREMENT , `application` VARCHAR(255) NOT NULL , `users` INT(11) NOT NULL , `is_active` TINYINT NOT NULL DEFAULT '1' , PRIMARY KEY (`id`)) ENGINE = InnoDB;"

	_, err := db.Exec(sqlStmt)

	if err != nil {
		log.Fatalln("Failed to create table, aborting")
	}

	log.Println(sqlStmt)
}
