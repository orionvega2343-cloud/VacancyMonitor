package main

import (
	"VacancyMonitor/internal/api"
	"VacancyMonitor/internal/config"
	"VacancyMonitor/internal/handlers"
	"VacancyMonitor/internal/models"
	"VacancyMonitor/internal/repository"
	"VacancyMonitor/internal/worker"
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	tele "gopkg.in/telebot.v3"
)

func main() {
	wg := &sync.WaitGroup{}
	workerCount := 5
	pollInterval := 5 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Конфиг
	cfg := config.MustLoad()

	//Общий канал
	share := make(chan models.Vacancy)

	//Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Host,
	})

	//Репозиторий
	repo := repository.NewVacancyRepo(rdb, "seen_vacancies")

	//Бот
	pref := tele.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	//Обработчик
	hndlr := handlers.NewVacancyHandler(b, share)

	//fetcher
	fetcher := api.NewVacancyFetcher(&http.Client{}, "VacancyMonitor/1.0 (ao0004@mail.ru)")

	//Rate-limiter
	limiter := api.NewRateLimiter(rdb, "hh_api_rate_limit", 10, time.Minute)

	//worker
	w := worker.NewWorker(fetcher, repo, wg, workerCount, models.Filter{}, pollInterval, share, limiter)

	hndlr.HandleFunc()
	go w.Call(ctx)
	go hndlr.Listen()
	b.Start()
}
