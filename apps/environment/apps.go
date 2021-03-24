package environment

import (
	"log"

	"github.com/lbryio/lbrytv/config"
	"github.com/lbryio/lbrytv/internal/storage"
)

func ForCollector() *Environment {
	e := &Environment{objects: map[string]interface{}{}}

	appConfig := config.ReadConfig("collector")
	dbConfig := appConfig.GetDatabase()
	conn := storage.InitConn(storage.ConnParams{
		Connection:     dbConfig.Connection,
		DBName:         dbConfig.DBName,
		Options:        dbConfig.Options,
		MigrationsPath: "/apps/collector/migrations",
	})

	err := conn.Connect()
	if err != nil {
		panic(err)
	}

	e.Add("config", appConfig)
	e.Add("storage", conn)

	return e
}

func ForWatchman() *Environment {
	e := &Environment{objects: map[string]interface{}{}}

	appConfig := config.ReadConfig("collector")
	dbConfig := appConfig.GetDatabase()
	conn := storage.InitConn(storage.ConnParams{
		Connection:     dbConfig.Connection,
		DBName:         dbConfig.DBName,
		Options:        dbConfig.Options,
		MigrationsPath: "/apps/watchman/migrations",
	})

	err := conn.Connect()
	if err != nil {
		log.Fatal("db connection failed:", err)
	}

	e.Add("config", appConfig)
	e.Add("storage", conn)

	return e
}
