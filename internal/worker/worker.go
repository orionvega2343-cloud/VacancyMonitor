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
	vacancyChan  chan models.Vacancy
	fetcher      api.VacancyFetcher
	repo         repository.VacancyRepo
	res          chan models.Vacancy
	wg           *sync.WaitGroup
	workerCount  int
	filter       models.Filter
	pollInterval time.Duration
	share        chan models.Vacancy
}

func NewWorker(vacancyChan chan models.Vacancy, fetcher api.VacancyFetcher, repo repository.VacancyRepo, res chan models.Vacancy, wg *sync.WaitGroup, workerCount int, filter models.Filter, pollInterval time.Duration, share chan models.Vacancy) *Worker {
	return &Worker{vacancyChan: vacancyChan, fetcher: fetcher, repo: repo, res: res, wg: wg, workerCount: workerCount, filter: filter, pollInterval: pollInterval, share: share}
}

//Создаем данные и
//раздаем их воркерам

func (w *Worker) Producer(ctx context.Context, filter models.Filter) <-chan models.Vacancy {
	out := w.vacancyChan
	go func() {
		defer close(out)
		vac, err := w.fetcher.FetchVacancies(ctx, filter)
		if err != nil {
			log.Println(err)
			return
		}
		for _, v := range vac {
			out <- v
		}

	}()

	return out
}

func (w *Worker) Worker(ctx context.Context, ch <-chan models.Vacancy) <-chan models.Vacancy {
	out := w.res

	go func() {
		defer w.wg.Done()
		for n := range ch {
			duplicate, err := w.repo.IsDuplicate(ctx, n.Id)
			if err != nil {
				log.Println(err)
				continue
			}
			if !duplicate {
				out <- n
			}
		}

	}()
	return out
}

//Добавляем количество воркеров к счетчику и
//ждем когда воркеры закончат работу и
//закрываем канал

func (w *Worker) Merge() <-chan models.Vacancy {
	out := w.res

	w.wg.Add(w.workerCount)

	go func() {
		w.wg.Wait()
		close(out)
	}()
	return out
}

func (w *Worker) call(ctx context.Context) {
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
				go w.Worker(ctx, in)
			}
			for v := range out {
				w.share <- v
			}
		}
	}
}
