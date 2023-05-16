package RSS

import (
	"encoding/json"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/mmcdole/gofeed"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.uber.org/zap"
	"time"
)

var (
	logger = logging.GetLogger()
	cfg    = config.GetConfig()
	parser = gofeed.NewParser()
)

// Startup - запускает cron job
// для парсинга источников из postgres.
// Время работы определяется в app.yml
func (u *Usecase) Startup() {
	s := gocron.NewScheduler(time.UTC)
	logger.Info("Парсим каждые " + cfg.CronSourcesRSS)

	_, err := s.Cron(cfg.CronSourcesRSS).Do(func() {
		for i, source := range sources {
			logger.Info("Парсим источинк #" + string(rune(i)))
			u.ParseRSS(source)
		}
	})
	if err != nil {
		logger.Fatal("Не стартовали CRON RSS", zap.Error(err))
	}
	s.StartAsync()
}

/*
TODO: [PROS-29] подключаю RSS
3 база данных источников
4 сохранение статей
*/

// Service - зависимые сервисы
type services interface {
	// Startup запускает постоянный RSS парсер
	Startup()
	// AddSourceRSS добавляет RSS ссылку
	AddSourceRSS(src string)
	// ParseRSS достаёт контент по источнику
	ParseRSS(src string)
}

// Usecase использование сервисов
type Usecase struct {
	services
}

// TODO: заменить на postgres
var sources = []string{
	"https://rss.nytimes.com/services/xml/rss/nyt/World.xml",
}

func (u *Usecase) AddSource(src string) {
	sources = append(sources, src)
}

func (u *Usecase) ParseRSS(src string) {
	feed, _ := parser.ParseURL(src)
	article, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		logger.Error("Не смогли распарсить источник", zap.Error(err))
	} else {
		logger.Info(fmt.Sprintf("Обработали: %s {%d}", src, len(string(article))))
		//for _, s := range strings.Split(string(article), "\n") {
		//	fmt.Println(s)
		//}
	}
}
