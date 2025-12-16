package handlers

import (
	"encoding/json"
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

func (h *PoolHandler) AddPool(c *gin.Context) {

	var body struct {
		Name string `json:"name" binding:"required"`
		Category string `json:"category" binding:"required"`
		Description string `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name, category and description are required"})
		return
	}

	payload := map[string]string{
		"name": 		body.Name,
		"category":     body.Category,
		"description": 	body.Description,
	}
	reqBytes, _ := json.Marshal(payload)

	msg, err := h.NatsConn.Request("pool.add", reqBytes, 2*time.Second)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"Error": "Worker unavailable"})
		return
	}

	c.Data(http.StatusOK, "application/json", msg.Data)
}