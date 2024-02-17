package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/config"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/storage/memory"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/storage/postgresql/postgresql_ast"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/storage/postgresql/postgresql_config"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/tasks/arithmetic"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/tasks/queue"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/tasks/responseStruct"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/validator"
)

func NewServeMux(config *config.Config,
	queue *queue.MapQueue,
	storage *memory.Storage,
) (http.Handler, error) {
	// Создам маршрутизатор
	serveMux := http.NewServeMux()
	// Регистрируем обработчики событий
	patchToFront := "./frontend/build"
	serveMux.Handle("/", http.FileServer(http.Dir(patchToFront)))
	serveMux.HandleFunc("/duration", durationHandler(config))
	serveMux.HandleFunc("/expression", expressionHandler(config, queue, storage))
	serveMux.HandleFunc("/id/", getById)
	serveMux.HandleFunc("/workers", getWorkers(config))
	return serveMux, nil
}

func Decorate(next http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	decorated := next
	for i := len(middleware) - 1; i >= 0; i-- {
		decorated = middleware[i](decorated)
	}

	return decorated
}

func getById(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
}

func getWorkers(conf *config.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			newWorkers := config.Workers{
				Agents:      conf.AgentsAll.Load(),
				Workers:     conf.WorkersAll.Load(),
				WorkersBusy: conf.WorkersBusy.Load(),
			}
			data, err := json.Marshal(newWorkers)
			if err != nil {
				slog.Error("Невозможно сериализовать данные:", "ошибка:", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.Write(data)
		}
	}
}

func expressionHandler(config *config.Config,
	queue *queue.MapQueue,
	storage *memory.Storage,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			type Expression interface {
				Result() []string
			}
			data, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("Проблема с чтением данных:", "ошибка:", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			slog.Info("Полученное выражение от пользователя:", "выражение:", string(data))
			// Валидируем входящее выражение
			str, ok := validator.Validator(string(data))
			if !ok {
				slog.Error("Некорректное выражение:", "ошибка:", err)
				http.Error(w, "Ваше выражение "+str+" некорректное", http.StatusBadRequest)
				return
			}
			// Проверяем, есть ли такое выражение в базе. Если есть - отдаем
			dataInfo, err := storage.GeByExpression(str)
			if err == nil {
				resp := responseStruct.NewExpression(dataInfo.Expression)
				data, err := json.Marshal(resp)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				w.WriteHeader(http.StatusAccepted)
				w.Write(data)
				slog.Info("Такое выражение уже было в базе", "ответ:", string(data))
				return
			}
			// Формируем новое выражение для вычисления
			exp, err := arithmetic.NewASTTree(str, config, queue)
			if err != nil {
				slog.Error("Проблема с вычислением выражения:", "выражение:", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// Сохраняем в память
			storage.Set(exp, "new")
			postgresql_ast.Add(exp, config)
			resp := responseStruct.NewExpression(exp)
			answer, err := json.Marshal(resp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.WriteHeader(http.StatusAccepted)
			w.Write(answer)
			slog.Info("Выражение добавлено в базу", "ответ:", string(answer))
		}
	}
}

func durationHandler(conf *config.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			newDuration := config.ConfigExpression{}
			data, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("Проблема с чтением данных:", "ошибка:", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			slog.Info("Полученное время для операций от пользователя:", "данные:", string(data))
			json.Unmarshal(data, &newDuration)
			err = conf.NewDuration(&newDuration)
			if err != nil {
				slog.Error("Некорректное время:", "ошибка:", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data, err = json.Marshal(newDuration)
			if err != nil {
				slog.Error("Невозможно сериализовать данные:", "ошибка:", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// Сохраняем в базу
			postgresql_config.Save(conf)
			slog.Info("Время для операций обновлено и отправлено", "новое время:", newDuration)
			w.Write(data)
		}
		if r.Method == http.MethodGet {
			newDuration := config.ConfigExpression{}
			newDuration.Init(conf)
			data, err := json.Marshal(newDuration)
			if err != nil {
				slog.Error("Невозможно сериализовать данные:", "ошибка:", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.Write(data)
		}
	}
}
