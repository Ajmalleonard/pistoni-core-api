package handlers

import "github.com/gin-gonic/gin"

type Handlers struct{}

func New() *Handlers { return &Handlers{} }

func (h *Handlers) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "service": "pistoni-core-go"})
}

func (h *Handlers) NotImplemented(route string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{
			"message": "endpoint scaffolded; implementation in progress",
			"route":   route,
		})
	}
}
