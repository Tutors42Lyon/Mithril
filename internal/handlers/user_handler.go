package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Tutors42Lyon/Mithril/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

type UserHandler struct {
	NatsConn *nats.Conn
}

func NewUserHandler(nc *nats.Conn) *UserHandler {
	return &UserHandler{
		NatsConn: nc,
	}
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	username := c.Param("username")

	var body struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role is required"})
		return
	}

	payload := map[string]string{
		"username": username,
		"role":     body.Role,
	}
	reqBytes, _ := json.Marshal(payload)

	msg, err := h.NatsConn.Request("user.update_role", reqBytes, 2*time.Second)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Worker unavailable"})
		return
	}

	var respMsg models.RespondMessage
    if err := json.Unmarshal(msg.Data, &respMsg); err != nil {
         c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response from worker"})
         return
    }

    if respMsg.HTTPStatus >= 400 {
        c.Data(respMsg.HTTPStatus, "application/json", msg.Data)
        return
    }

	c.Data(http.StatusOK, "application/json", msg.Data)

}

func (h *UserHandler) GetUserInfo(c *gin.Context) {

	username := c.Param("username")

	payload := map[string]string{"username": username}
	reqBytes, _ := json.Marshal(payload)

	msg, err := h.NatsConn.Request("user.get_info", reqBytes, 2*time.Second)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"Error": "Worker unavailable"})
		return
	}

	var respMsg models.RespondMessage
    if err := json.Unmarshal(msg.Data, &respMsg); err != nil {
         c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response from worker"})
         return
    }

	if respMsg.HTTPStatus >= 400 {
        c.Data(respMsg.HTTPStatus, "application/json", msg.Data)
        return
    }
	c.Data(http.StatusOK, "application/json", msg.Data)
}
