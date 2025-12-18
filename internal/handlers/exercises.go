package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

type ExerciseHandler struct {
	NatsConn *nats.Conn
}

func NewExerciseHandler(nc *nats.Conn) *ExerciseHandler {
	return &ExerciseHandler{
		NatsConn: nc,
	}
}

func (h *ExerciseHandler) GetExercisesFromPool(c *gin.Context) {

	poolName := c.Param("poolname")

	payload := map[string]string{"poolname": poolName}
	reqBytes, _ := json.Marshal(payload)

	msg, err := h.NatsConn.Request("exercises.get_info", reqBytes, 2*time.Second)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"Error": "Worker unavailable"})
		return
	}

	c.Data(http.StatusOK, "application/json", msg.Data)
}