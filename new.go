package ginplus

import "github.com/gin-gonic/gin"

type (
	GinEngine struct {
		*gin.Engine
		Controllers []Controller
	}

	Controller interface {
		Middlewares() []gin.HandlerFunc
	}

	Route struct {
		Path       string
		HttpMethod string
		Handles    []gin.HandlerFunc
	}

	Option func(*GinEngine)
)

// New returns a GinEngine instance.
func New(r *gin.Engine, opts ...Option) *GinEngine {
	instance := &GinEngine{Engine: r}
	for _, opt := range opts {
		opt(instance)
	}

	routes := make([]*Route, 0)
	for _, c := range instance.Controllers {
		routes = append(routes, genRoute(c)...)
	}

	for _, route := range routes {
		instance.Handle(route.HttpMethod, route.Path, route.Handles...)
	}

	return instance
}

// WithControllers sets the controllers.
func WithControllers(controllers ...Controller) Option {
	return func(g *GinEngine) {
		g.Controllers = controllers
	}
}
