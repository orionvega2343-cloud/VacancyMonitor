package handlers

import (
	"VacancyMonitor/internal/models"
	"fmt"
	"log"
	"os"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

type VacancyHandler struct {
	b     *tele.Bot
	share <-chan models.Vacancy
}

func NewVacancyHandler(b *tele.Bot, share <-chan models.Vacancy) *VacancyHandler {
	return &VacancyHandler{b: b, share: share}
}

func (h *VacancyHandler) HandleFunc() {
	h.b.Handle("/start", func(c tele.Context) error {
		return c.Send("Список вакансий:")
	})
}

func (h *VacancyHandler) Listen() {
	for v := range h.share {
		text := fmt.Sprintf("Вакансия: %s, ссылка: %s, зарплата: от %d до %d %s (%v), регион: %s, уровень опыта: %s", v.Name, v.AlternateUrl, v.Salary.From, v.Salary.To, v.Salary.Currency, v.Salary.Gross, v.Area, v.Experience)
		id := os.Getenv("chat_id")
		parsed, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Println(err)
			continue
		}
		_, err = h.b.Send(tele.ChatID(parsed), text)
		if err != nil {
			log.Println(err)
		}
	}
}
