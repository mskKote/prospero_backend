package main

import (
	"context"
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/internal/adapters/db/elastic/articlesSearchRepository"
	publishersSearchRepository "github.com/mskKote/prospero_backend/internal/adapters/db/elastic/publisherSearchRepository"
	"github.com/mskKote/prospero_backend/internal/adapters/db/postgres/adminsRepository"
	"github.com/mskKote/prospero_backend/internal/adapters/db/postgres/publishersRepository"
	"github.com/mskKote/prospero_backend/internal/adapters/db/postgres/sourcesRepository"
	internalMetrics "github.com/mskKote/prospero_backend/internal/adapters/metrics"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/routes"
	"github.com/mskKote/prospero_backend/internal/domain/entity/admin"
	"github.com/mskKote/prospero_backend/internal/domain/service/adminService"
	"github.com/mskKote/prospero_backend/internal/domain/service/articleService"
	"github.com/mskKote/prospero_backend/internal/domain/service/publishersService"
	"github.com/mskKote/prospero_backend/internal/domain/service/sourcesService"
	"github.com/mskKote/prospero_backend/internal/domain/usecase/RSS"
	"github.com/mskKote/prospero_backend/internal/domain/usecase/adminka"
	"github.com/mskKote/prospero_backend/internal/domain/usecase/search"
	"github.com/mskKote/prospero_backend/pkg/client/elastic"
	"github.com/mskKote/prospero_backend/pkg/client/postgres"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	pkgMetrics "github.com/mskKote/prospero_backend/pkg/metrics"
	"github.com/mskKote/prospero_backend/pkg/security"
	"github.com/mskKote/prospero_backend/pkg/tracing"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	cfg    = config.GetConfig()
	logger = logging.GetLogger()
)

func main() {
	startup(cfg)
}

func startup(cfg *config.Config) {

	// --------------------------------------- DATABASES
	ctx := context.Background()
	pgClient, err := postgres.NewClient(ctx, 3)
	if err != nil {
		logger.Fatal("[POSTGRES] Не подключились к postgres", zap.Error(err))
	} else {
		logger.Info("[POSTGRES] УСПЕШНО подключилсь к POSTGRES!")
	}

	esClient, err := elastic.NewClient(ctx)
	if err != nil {
		logger.Fatal("[ELASTIC] Не подключились к elastic", zap.Error(err))
	} else {
		logger.Info("[ELASTIC] УСПЕШНО подключилсь к ELASTICSEARCH!")
	}

	if cfg.MigratePostgres {
		migrationsPg(pgClient, ctx)
	}
	if cfg.MigrateElastic {
		migrationsEs(esClient, ctx)
	}

	sourcesREPO := sourcesRepository.New(pgClient)
	articlesREPO := articlesSearchRepository.New(esClient)
	publishersREPO := publishersRepository.New(pgClient)
	publishersSearchREPO := publishersSearchRepository.New(esClient)

	publishersSERVICE := publishersService.New(publishersREPO, publishersSearchREPO)
	articlesSERVICE := articleService.New(sourcesREPO, articlesREPO)
	sourcesSERVICE := sourcesService.New(sourcesREPO)

	// --------------------------------------- GIN
	r := gin.New()
	if cfg.IsDebug == false {
		gin.SetMode(gin.ReleaseMode)
	}

	corsCfg := cors.DefaultConfig()
	corsCfg.AllowAllOrigins = true
	corsCfg.AddExposeHeaders(tracing.ProsperoHeader)
	corsCfg.AddAllowHeaders("Authorization")
	r.Use(cors.New(corsCfg))

	// массив из cfg?
	//err := r.SetTrustedProxies([]string{"127.0.0.1"})
	//if err != nil {
	//	logger.Fatal("Не получилось установить proxy", zap.Error(err))
	//}

	// Recovery
	r.Use(gin.Recovery())

	// Logging
	if cfg.Logger.UseDefaultGin {
		logger.Info("Используем DefaultGin")
		r.Use(gin.Logger())
	}
	//if cfg.Logger.ToGraylog { // logrus & graylog
	//	logger.Info("Используем Graylog")
	//	r.Use(logging.GraylogMiddlewareLogger())
	//}
	if cfg.Logger.UseZap {
		logger.Info("Используем Zap")
		logging.ZapMiddlewareLogger(r)
		undo := otelzap.ReplaceGlobals(logger.Logger)
		defer undo()
		defer func(loggerZap *zap.Logger) {
			err := loggerZap.Sync()
			if err != nil {
				loggerZap.Error("[LOGGER] Не получилось синхронизироваться", zap.Error(err))
			}
		}(logger.Logger.Logger)
	}

	// Tracing
	if cfg.UseTracingJaeger {
		tp := tracing.Startup(r)
		ctx, cancel := context.WithCancel(context.Background())

		// Cleanly shutdown and flush telemetry when the application exits.
		defer func(ctx context.Context) {
			// Do not make the application hang when it is shutdown.
			ctx, cancel = context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err := tp.Shutdown(ctx); err != nil {
				logger.Fatal("[TRACING] Ошибка при выключении", zap.Error(err))
			}
		}(ctx)
	}

	// Metrics
	if cfg.Metrics {
		p := pkgMetrics.Startup(r)
		internalMetrics.RegisterMetrics(p)
	}

	// --------------------------------------- ROUTES
	prosperoRoutes(r, &publishersSERVICE, &articlesSERVICE)
	adminkaStartup(r, pgClient, &sourcesSERVICE, &publishersSERVICE, &articlesSERVICE)

	// --------------------------------------- IGNITION
	if cfg.UseCronSourcesRSS {
		go RSS.New(sourcesSERVICE, articlesSERVICE).Startup()
	}

	if err := r.Run(":" + cfg.Port); err != nil {
		logger.Fatal("ошибка, завершаем программу", zap.Error(err))
	}
}

func migrationsPg(client postgres.Client, ctx context.Context) {
	migration, err := os.OpenFile("./resources/migration_20230517_1.sql", os.O_RDONLY, 0666)
	if err != nil {
		logger.Fatal("[MIGRATION] Невозможно прочитать файл", zap.Error(err))
	}
	defer func(migration *os.File) {
		err := migration.Close()
		if err != nil {
			logger.Fatal("[MIGRATION] Невозможно закрыть файл миграции", zap.Error(err))
		}
	}(migration)

	data, err := io.ReadAll(migration)
	_, err = client.Exec(ctx, string(data))
	if err != nil {
		logger.Fatal("[MIGRATION] Миграции POSTGRES провалились", zap.Error(err))
	} else {
		logger.Info("[MIGRATION] УСПЕШНО мигрировали POSTGRES")
	}
}

func migrationsEs(client *elasticsearch.TypedClient, ctx context.Context) {
	log.Printf("\n\n")
	publishersSearchRepository.New(client).Setup(ctx)
	articlesSearchRepository.New(client).Setup(ctx)
}

func adminkaStartup(
	r *gin.Engine,
	client postgres.Client,
	s *sourcesService.ISourceService,
	p *publishersService.IPublishersService,
	a *articleService.IArticleService) {

	adminREPO := adminsRepository.New(client)
	adminSERVICE := adminService.New(adminREPO)
	adminkaUSECASE := adminka.New(s, p, a)

	// Админ
	if cfg.MigratePostgres {
		adminMskKote := &admin.DTO{
			Name:     cfg.Adminka.Username,
			Password: cfg.Adminka.Password,
		}
		logger.Info(fmt.Sprintf("[ADMINKA] Админка: {%s}, {%s}", adminMskKote.Name, adminMskKote.Password))

		if err := adminSERVICE.Create(context.Background(), adminMskKote); err != nil {
			logger.Fatal("[ADMINKA] Не смогли создать админа: "+adminMskKote.Name, zap.Error(err))
		} else {
			logger.Info(fmt.Sprintf("[ADMINKA] Админка: {%s}, {%s}", adminMskKote.Name, adminMskKote.Password))
		}
	}

	auth := security.Startup(adminSERVICE)

	adminkaGroup := r.Group("/adminka")
	adminkaGroup.POST("/login", auth.LoginHandler)
	adminkaGroup.OPTIONS("/login")
	adminkaGroup.Use(auth.MiddlewareFunc())
	{
		adminkaGroup.GET("/refresh_token", auth.RefreshHandler)

		// TEST STAND
		adminkaGroup.GET("/hello", func(c *gin.Context) {
			claims := jwt.ExtractClaims(c)
			if user, ok := c.Get("id"); ok {
				c.JSON(http.StatusOK, gin.H{
					"userID":   claims["id"],
					"userName": user.(*admin.Admin).Name,
					"text":     "Hello World.",
				})
			} else {
				c.JSON(http.StatusNotFound, gin.H{
					"userID":   claims["id"],
					"userName": "Not found",
					"text":     "Bye World.",
				})
			}
		})

		adminkaApiV1 := adminkaGroup.Group("api/v1")
		routes.RegisterSourcesRoutes(adminkaApiV1, adminkaUSECASE)
		routes.RegisterPublishersRoutes(adminkaApiV1, adminkaUSECASE)
	}

	r.NoRoute(auth.MiddlewareFunc(), security.NoRoute)
}

func prosperoRoutes(
	r *gin.Engine,
	p *publishersService.IPublishersService,
	a *articleService.IArticleService) {

	searchUSECASE := search.New(p, a)

	apiV1 := r.Group("/api/v1")
	{
		routes.RegisterSearchRoutes(apiV1, searchUSECASE)
	}
}
