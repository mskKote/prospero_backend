package main

import (
	"context"
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/internal/adapters/db/postgres/adminsRepository"
	"github.com/mskKote/prospero_backend/internal/adapters/db/postgres/publishersRepository"
	"github.com/mskKote/prospero_backend/internal/adapters/db/postgres/sourcesRepository"
	internalMetrics "github.com/mskKote/prospero_backend/internal/adapters/metrics"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/routes"
	"github.com/mskKote/prospero_backend/internal/domain/entity/admin"
	"github.com/mskKote/prospero_backend/internal/domain/service/adminService"
	"github.com/mskKote/prospero_backend/internal/domain/service/publishersService"
	"github.com/mskKote/prospero_backend/internal/domain/service/sourcesService"
	"github.com/mskKote/prospero_backend/internal/domain/usecase/RSS"
	"github.com/mskKote/prospero_backend/internal/domain/usecase/adminka"
	"github.com/mskKote/prospero_backend/internal/domain/usecase/search"
	"github.com/mskKote/prospero_backend/pkg/client/postgres"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	pkgMetrics "github.com/mskKote/prospero_backend/pkg/metrics"
	"github.com/mskKote/prospero_backend/pkg/security"
	"github.com/mskKote/prospero_backend/pkg/tracing"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"io"
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
	client, err := postgres.NewClient(context.Background(), 3)
	if err != nil {
		logger.Fatal("[POSTGRES] Не подключились к postgres", zap.Error(err))
	} else {
		logger.Info("[POSTGRES] УСПЕШНО подключилсь к POSTGRES!")
	}

	if cfg.Migrate {
		migrations(client)
	}

	// --------------------------------------- GIN
	r := gin.New()
	if cfg.IsDebug == false {
		gin.SetMode(gin.ReleaseMode)
	}
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
	if cfg.Tracing {
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
	prosperoRoutes(r)
	adminkaStartup(client, r)

	// --------------------------------------- IGNITION
	if cfg.UseCronSourcesRSS {
		go (&RSS.Usecase{}).Startup()
	}

	if err := r.Run(":" + cfg.Port); err != nil {
		logger.Fatal("ошибка, завершаем программу", zap.Error(err))
	}
}

func migrations(client postgres.Client) {
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
	_, err = client.Exec(context.Background(), string(data))
	if err != nil {
		logger.Fatal("[MIGRATION] Миграции POSTGRES провалились", zap.Error(err))
	} else {
		logger.Info("[MIGRATION] УСПЕШНО мигрировали POSTGRES")
	}
}

func adminkaStartup(client postgres.Client, r *gin.Engine) {
	adminREPO := adminsRepository.New(client)
	sourcesREPO := sourcesRepository.New(client)
	publishersREPO := publishersRepository.New(client)

	publishersSERVICE := publishersService.New(publishersREPO)
	sourcesSERVICE := sourcesService.New(sourcesREPO)
	adminSERVICE := adminService.New(adminREPO)

	adminkaUSECASE := adminka.New(sourcesSERVICE, publishersSERVICE)

	// Админ
	adminMskKote := &admin.DTO{
		Name:     cfg.Adminka.Username,
		Password: cfg.Adminka.Password,
	}
	if err := adminSERVICE.Create(context.Background(), adminMskKote); err != nil {
		logger.Fatal("[ADMINKA] Не смогли создать админа: "+adminMskKote.Name, zap.Error(err))
	} else {
		logger.Info(fmt.Sprintf("[ADMINKA] Админка: {%s}, {%s}", adminMskKote.Name, adminMskKote.Password))
	}

	auth := security.Startup(adminSERVICE)

	adminkaGroup := r.Group("/adminka")
	adminkaGroup.POST("/login", auth.LoginHandler)
	adminkaGroup.Use(auth.MiddlewareFunc())
	{
		adminkaGroup.GET("/refresh_token", auth.RefreshHandler)

		// TEST STAND
		adminkaGroup.GET("/hello", func(c *gin.Context) {
			claims := jwt.ExtractClaims(c)
			user, _ := c.Get("id")
			c.JSON(http.StatusOK, gin.H{
				"userID":   claims["id"],
				"userName": user.(*admin.Admin).Name,
				"text":     "Hello World.",
			})
		})

		adminkaApiV1 := adminkaGroup.Group("api/v1")
		routes.RegisterSourcesRoutes(adminkaApiV1, adminkaUSECASE)
		routes.RegisterPublishersRoutes(adminkaApiV1, adminkaUSECASE)
	}

	r.NoRoute(auth.MiddlewareFunc(), security.NoRoute)
}

func prosperoRoutes(r *gin.Engine) {
	searchUSECASE := &search.Usecase{}

	apiV1 := r.Group("/api/v1")
	{
		routes.RegisterSearchRoutes(apiV1, searchUSECASE)
	}
}
