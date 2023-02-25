package server

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/ksusonic/gophermart/internal/config"
	"go.uber.org/zap"
)

const apiPrefix = "/api"

type Server struct {
	Engine *gin.Engine
	logger *zap.SugaredLogger
}

func NewServer(cfg *config.Config, logger *zap.SugaredLogger) *Server {
	r := gin.Default()

	if cfg.Debug {
		logger.Debugf("loaded cfg: %s", cfg)
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r.Use(gzip.Gzip(gzip.DefaultCompression))

	_ = r.SetTrustedProxies([]string{})

	return &Server{
		Engine: r,
		logger: logger,
	}
}

type Controller interface {
	RegisterHandlers(routerGroup *gin.RouterGroup)
}

func (s *Server) MountController(path string, controller Controller) {
	controller.RegisterHandlers(s.Engine.Group(apiPrefix + path))
}

func (s *Server) Run(address string) error {
	s.logger.Infof("Starting server on %s", address)
	return s.Engine.Run(address)
}