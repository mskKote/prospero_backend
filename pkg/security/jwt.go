package security

import (
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var (
	logger      = logging.GetLogger()
	cfg         = config.GetConfig()
	identityKey = "id"
)

// AdminUser DTO
type AdminUser struct {
	UserID    string
	FirstName string
	LastName  string
}

// Startup - returns protected router group
func Startup() *jwt.GinJWTMiddleware {

	// the jwt middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "Prospero test zone",
		Key:         []byte(cfg.SecretKeyJWT),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*AdminUser); ok {
				return jwt.MapClaims{
					identityKey: v.UserID,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &AdminUser{
				UserID: claims[identityKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginValues login
			if err := c.ShouldBind(&loginValues); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			userID := loginValues.Username
			password := loginValues.Password

			//TODO: прокинуть функцию для обращение к базе админов
			if (userID == "admin" && password == "admin") ||
				(userID == "test" && password == "test") {
				return &AdminUser{
					UserID:    userID,
					LastName:  "Bo-Yi",
					FirstName: "Wu",
				}, nil
			}

			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			v, ok := data.(*AdminUser)
			return ok && v.UserID == "admin"
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})

	if err != nil {
		logger.Fatal("JWT Error", zap.Error(err))
	}

	// When you use jwt.New(), the function is already automatically called for checking,
	// which means you don't need to call it again.
	errInit := authMiddleware.MiddlewareInit()

	if errInit != nil {
		logger.Fatal("authMiddleware.MiddlewareInit", zap.Error(errInit))
	}

	return authMiddleware
}

func NoRoute(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	logger.Info(fmt.Sprintf("NoRoute claims: %#v\n", claims))
	c.JSON(http.StatusNotFound, gin.H{
		"code":    "PAGE_NOT_FOUND",
		"message": "Page not found",
	})
}
