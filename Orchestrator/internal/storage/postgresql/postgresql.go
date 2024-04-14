package postgresql

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/storage/postgresql/postgresql_ast"
	_ "github.com/lib/pq"
)

type Storage struct {
	Db        *sql.DB
	Extension *postgresql_ast.Data
}

// Создаем подключение к базе данных
func NewPostgresConnect(Db, DbPort, DbUser, DbPass, DbName string) *Storage {
	var db *sql.DB
	var err error
	connect := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		Db, DbPort, DbUser, DbPass, DbName)
	db, err = sql.Open("postgres", connect)
	if err != nil {
		slog.Error("Неверные данные для подключения к базе данных", "ОШИБКА:", err)
		panic(err)
	}
	return &Storage{Db: db, Extension: postgresql_ast.NewData(db)}
}
