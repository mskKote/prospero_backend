package adminka

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/routes"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
	"github.com/mskKote/prospero_backend/internal/domain/entity/source"
	"github.com/mskKote/prospero_backend/internal/domain/service/publishersService"
	"github.com/mskKote/prospero_backend/internal/domain/service/sourcesService"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"github.com/mskKote/prospero_backend/pkg/tracing"
	"go.uber.org/zap"
	"net/http"
)

var logger = logging.GetLogger()

// usecase - зависимые сервисы
type usecase struct {
	sources    sourcesService.ISourceService
	publishers publishersService.IPublishersService
}

type IAdminkaUsecase interface {
	routes.ISourcesUsecase
	routes.IPublishersUsecase
}

func New(
	s sourcesService.ISourceService,
	p publishersService.IPublishersService) IAdminkaUsecase {
	return &usecase{s, p}
}

// ---------------------------------------------------- sources CRUD

func (u *usecase) CreateSourceRSS(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)

	ctx := c.Request.Context()

	s := source.AddSourceDTO{}
	if err := c.Bind(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Неправильное тело запроса",
			"error":   err.Error()})
		return
	}

	if src, err := u.sources.AddSource(ctx, s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Не получилось добавить источник",
			"error":   err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
			"data":    src,
		})
	}
}

func (u *usecase) ReadSourcesRSS(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)

	ctx := c.Request.Context()
	dto := source.DTO{}
	if err := c.Bind(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Неправильное тело запроса",
			"error":   err.Error()})
		return
	}

	var sources []*source.DTO
	var err error

	if search := c.Query("search"); search != "" {
		sources, err = u.sources.FindByPublisherName(ctx, search)
	} else {
		sources, err = u.sources.FindAll(ctx)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не получилось найти RSS источники " + c.Query("search")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    sources,
	})
}

func (u *usecase) UpdateSourceRSS(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)

	ctx := c.Request.Context()
	dto := &source.DTO{}
	if err := c.Bind(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Неправильное тело запроса",
			"error":   err.Error()})
		return
	}

	if data, err := u.sources.Update(ctx, dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Не получилось обновить RSS источник",
			"error":   err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
			"data":    data,
		})
	}
}

func (u *usecase) DeleteSourceRSS(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)
	ctx := c.Request.Context()
	dto := source.DeleteSourceDTO{}
	if err := c.Bind(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Неправильное тело запроса",
			"error":   err.Error()})
		return
	}

	if err := u.sources.Delete(ctx, dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не получилось удалить RSS источник"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// ---------------------------------------------------- publishers CRUD

func (u *usecase) CreatePublisher(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)

	ctx := c.Request.Context()

	s := &publisher.AddPublisherDTO{}

	if err := c.Bind(&s); err != nil {
		logger.Error("Ошибка при добавлении источника", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Неправильное тело запроса",
			"error":   err.Error()})
		return
	}

	if data, err := u.publishers.Create(ctx, s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Ошибка при добавлении источника",
			"error":   err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
			"data":    data})
	}
}

func (u *usecase) ReadPublishers(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)

	ctx := c.Request.Context()
	dto := source.DTO{}
	if err := c.Bind(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Неправильное тело запроса",
			"error":   err.Error()})
		return
	}

	var publishers []*publisher.DTO
	var err error

	if search := c.Query("search"); search != "" {
		publishers, err = u.publishers.FindPublishersByName(ctx, search)
	} else {
		publishers, err = u.publishers.FindAll(ctx)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Не получилось найти RSS источники " + c.Query("search"),
			"error":   err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    publishers,
	})
}

func (u *usecase) UpdatePublisher(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)

	ctx := c.Request.Context()
	dto := &publisher.DTO{}
	if err := c.Bind(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Неправильное тело запроса",
			"error":   err.Error()})
		return
	}

	if err := u.publishers.Update(ctx, dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Не получилось обновить",
			"error":   err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (u *usecase) DeletePublisher(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)

	ctx := c.Request.Context()
	dto := &publisher.DeletePublisherDTO{}
	if err := c.Bind(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Неправильное тело запроса",
			"error":   err.Error()})
		return
	}

	if err := u.publishers.Delete(ctx, dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Не получилось удалить",
			"error":   err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
