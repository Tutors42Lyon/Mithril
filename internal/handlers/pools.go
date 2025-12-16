package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

type PoolHandler struct {
	NatsConn *nats.Conn
}

func NewPoolHandler(nc *nats.Conn) *PoolHandler {
	return &PoolHandler{
		NatsConn: nc,
	}
}

func (h *PoolHandler) GetPoolsInfo(c *gin.Context) {
	msg, err := h.NatsConn.Request("pools.get_info", nil, 2*time.Second)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"Error": "Worker unavailable"})
		return
	}

	c.Data(http.StatusOK, "application/json", msg.Data)
}