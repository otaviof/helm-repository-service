package chartstreams

import (
	"github.com/gin-gonic/gin"
)

// ChartStreamServer represents the chart-streams server offering its API. The server puts together the routes,
// and bootstrap steps in order to respond as a valid Helm repository.
type ChartStreamServer struct {
	config     *Config
	gitService *GitService
}

// Start executes the boostrap steps in order to start listening on configured address. It can return
// errors from "listen" method.
func (s *ChartStreamServer) Start() error {
	if err := s.gitService.Initialize(); err != nil {
		return err
	}

	return s.listen()
}

func (s *ChartStreamServer) IndexHandler(c *gin.Context) {
	index, err := s.gitService.GetIndex()
	if err != nil {
		c.AbortWithError(500, err)
	}

	c.YAML(200, index)
}

func (s *ChartStreamServer) DirectLinkHandler(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")

	err := s.gitService.GetHelmChart(name, version)
	if err != nil {
		c.AbortWithError(500, err)
	}
}

// listen on configured address, after adding the route handlers to the framework. It can return
// errors coming from Gin.
func (s *ChartStreamServer) listen() error {
	g := gin.Default()

	g.GET("/index.yaml", s.IndexHandler)
	g.GET("/chart/:name/*version", s.DirectLinkHandler)

	return g.Run(s.config.ListenAddr)
}

// NewServer instantiate a new server instance.
func NewServer(config *Config) *ChartStreamServer {
	gs := NewGitService(config)
	return &ChartStreamServer{
		config:     config,
		gitService: gs,
	}
}
