package service

import (
	"time"

	"github.com/jmoloko/taskmange/internal/domain/models"
)

// Analytics представляет аналитические данные по задачам для кэширования
type Analytics struct {
	StatusCount          map[models.Status]int
	PriorityCount        map[models.Priority]int
	AvgCompletionTime    float64
	OnTimeCompletionRate float64
	OverdueTasks         int
	Period               string
	GeneratedAt          time.Time
}
