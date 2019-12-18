package helmrepositoryservice

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/redhat-developer/helm-repository-service/pkg/helmrepositoryservice/config"
	"github.com/redhat-developer/helm-repository-service/pkg/helmrepositoryservice/provider"
)

// ChartStreamServer represents the chartstreams server offering its API. The server puts together
// the routes, and bootstrap steps in order to respond as a valid Helm repository.
type ChartStreamServer struct {
	config        *config.Config
	chartProvider provider.ChartProvider
}

// Start executes the boostrap steps in order to start listening on configured address. It can return
// errors from "listen" method.
func (s *ChartStreamServer) Start() error {
	if err := s.chartProvider.Initialize(); err != nil {
		return err
	}

	return s.listen()
}

// IndexHandler endpoint handler to render a index.yaml file.
func (s *ChartStreamServer) IndexHandler(c *gin.Context) {
	index, err := s.chartProvider.GetIndexFile()
	if err != nil {
		c.AbortWithError(500, err)
	}

	c.YAML(200, index)
}

// DirectLinkHandler endpoint handler to directly load a chart tarball payload.
func (s *ChartStreamServer) DirectLinkHandler(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	version = strings.TrimPrefix(version, "/")

	p, err := s.chartProvider.GetChart(name, version)
	if err != nil {
		c.AbortWithError(500, err)
	}

	c.Data(http.StatusOK, "application/gzip", p.Bytes())
}

// listen on configured address, after adding the route handlers to the framework. It can return
// errors coming from Gin.
func (s *ChartStreamServer) listen() error {
	g := gin.New()

	g.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))

	g.GET("/index.yaml", s.IndexHandler)
	g.GET("/chart/:name/*version", s.DirectLinkHandler)

	return g.Run(s.config.ListenAddr)
}

// NewServer instantiate a new server instance.
func NewServer(config *config.Config) *ChartStreamServer {
	p := provider.NewGitChartProvider(config)
	return &ChartStreamServer{
		config:        config,
		chartProvider: p,
	}
}
