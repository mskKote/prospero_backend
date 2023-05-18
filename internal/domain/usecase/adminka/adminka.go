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
	"strconv"
)

var logger = logging.GetLogger()

const pageSize int = 2

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
	var src []*source.DTO
	var err error

	// Total
	var total int64
	if count, err := u.sources.Count(ctx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Не получилось посчитать количество элементов " + c.Query("search"),
			"error":   err.Error()})
		return
	} else {
		total = count
	}

	pageQuery := c.DefaultQuery("page", "0")
	page, err := strconv.Atoi(pageQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Неправильные параметры запроса",
			"error":   err.Error()})
		return
	}

	if int(total) < pageSize*page {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Страницы " + c.Query("page") + " нет",
			"error":   err.Error()})
		return
	}

	if search := c.Query("search"); search != "" {
		src, err = u.sources.FindByPublisherName(ctx, search, page, pageSize)
	} else {
		src, err = u.sources.FindAll(ctx, page, pageSize)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не получилось найти RSS источники " + c.Query("search")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    src,
	})
}

func (u *usecase) ReadSourcesRSSWithPublishers(c *gin.Context) {
	tracing.TraceHeader(c)
	tracing.LogRequestTrace(c)

	ctx := c.Request.Context()
	var src []*source.DTO
	var err error

	// Total
	var total int64
	if count, err := u.sources.Count(ctx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Не получилось посчитать количество элементов " + c.Query("search"),
			"error":   err.Error()})
		return
	} else {
		total = count
	}

	pageQuery := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Неправильные параметры запроса",
			"error":   err.Error()})
		return
	}

	if int(total) < pageSize*page {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Страницы " + c.Query("page") + " нет"})
		return
	}

	// Поиск
	if search := c.Query("search"); search != "" {
		src, err = u.sources.FindByPublisherName(ctx, search, page-1, pageSize)
	} else {
		src, err = u.sources.FindAll(ctx, page-1, pageSize)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Не получилось найти RSS источники " + c.Query("search"),
			"error":   err.Error()})
		return
	}

	// Enrich
	pubMap := map[string]*publisher.DTO{}

	var srcIDs []string
	for _, s := range src {
		srcIDs = append(srcIDs, s.PublisherID)
	}

	if publishers, err := u.publishers.FindPublishersByIDs(ctx, srcIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Не нашли publishers",
			"error":   err.Error()})
		return
	} else {
		for _, p := range publishers {
			pubMap[p.PublisherID] = p
		}
	}

	// Join 2 сущностей по publisherID
	var srcEnriched []*source.WithPublisher
	for _, s := range src {
		if p, ok := pubMap[s.PublisherID]; ok {
			srcEnriched = append(srcEnriched, &source.WithPublisher{
				Source:    s,
				Publisher: p,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    srcEnriched,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
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
