package main

const (
	dateFormat = "20060102"
)

const (
	SCHEMA = `CREATE TABLE IF NOT EXISTS "scheduler" (
					id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
					date CHAR(8) NOT NULL,
					title TEXT NOT NULL,
					comment TEXT,
					repeat VARCHAR(128)
			);`
	DB_NAME              = "scheduler.db"
	DB_FILE              = "./" + DB_NAME
	DATE_PATTERN         = "[0-9]{2}\\.[0-9]{2}\\.[12]{1}[0-9]{3}"
	DATE_SCHEDULE_FORMAT = "20060102"
	DATE_SEARCH_FORMAT   = "02.01.2006"
	LIMIT                = 50
)
