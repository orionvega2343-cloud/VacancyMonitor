package worker

import (
	"VacancyMonitor/internal/api"
	"VacancyMonitor/internal/models"
	"VacancyMonitor/internal/repository"
	"context"
	"log"
	"sync"
	"time"
)

type Worker struct {
	fetcher      api.VacancyFetcher
	repo         repository.VacancyRepo
	wg           *sync.WaitGroup
	workerCount  int
	filter       models.Filter
	pollInterval time.Duration
	share        chan models.Vacancy
	limiter      api.RateLimiter
}

func NewWorker(fetcher api.VacancyFetcher, repo repository.VacancyRepo, wg *sync.WaitGroup, workerCount int, filter models.Filter, pollInterval time.Duration, share chan models.Vacancy, limiter api.RateLimiter) *Worker {
	return &Worker{fetcher: fetcher, repo: repo, wg: wg, workerCount: workerCount, filter: filter, pollInterval: pollInterval, share: share, limiter: limiter}
}

//Создаем данные и
//раздаем их воркерам

func (w *Worker) Producer(ctx context.Context, filter models.Filter) <-chan models.Vacancy {
	out := make(chan models.Vacancy)
	go func() {
		defer close(out)

		err := w.limiter.Wait(ctx)
		if err != nil {
			log.Println(err)
			return
		}

		vac, err := w.fetcher.FetchVacancies(ctx, filter)
		if err != nil {
			log.Println(err)
			return
		}
		for _, v := range vac {
			//Пытаемся отправить значение в канал;
			//если контекст отменён раньше, чем получится отправить — выходим
			select {
			case out <- v:
			case <-ctx.Done():
				return
			}
		}

	}()

	return out
}

func (w *Worker) Worker(ctx context.Context, ch <-chan models.Vacancy, out chan<- models.Vacancy) {
	defer w.wg.Done()
	for n := range ch {
		duplicate, err := w.repo.IsDuplicate(ctx, n.Id)
		if err != nil {
			log.Println(err)
			continue
		}
		if !duplicate {
			//Пытаемся отправить значение в канал;
			//если контекст отменён раньше, чем получится отправить — выходим
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}
}

//Добавляем количество воркеров к счетчику и
//ждем когда воркеры закончат работу и
//закрываем канал

func (w *Worker) Merge() chan models.Vacancy {
	out := make(chan models.Vacancy)

	w.wg.Add(w.workerCount)

	go func() {
		w.wg.Wait()
		close(out)
	}()
	return out
}

func (w *Worker) Call(ctx context.Context) {
	//тикер непрерывного вызова producer
	//для слежки за новыми вакансиями
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		//При завершении работы контекста,
		//полностью выходим с функции
		case <-ctx.Done():
			return
		case <-ticker.C:
			in := w.Producer(ctx, w.filter)
			out := w.Merge()
			for i := 0; i < w.workerCount; i++ {
				go w.Worker(ctx, in, out)
			}
			for v := range out {
				select {
				case w.share <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
