package server_handler

import (
	"github.com/evolution-foundation/evolution-go/pkg/dbstats"
	"github.com/gin-gonic/gin"
)

type ServerHandler interface {
	ServerOk(ctx *gin.Context)
	DBStats(ctx *gin.Context)
}

type serverHandler struct {
}

// ServerOk implements ServerHandler.
func (s *serverHandler) ServerOk(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status": "ok",
	})
}

func (s *serverHandler) DBStats(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"pools": dbstats.Snapshot(),
	})
}

func NewServerHandler() ServerHandler {
	return &serverHandler{}
}
