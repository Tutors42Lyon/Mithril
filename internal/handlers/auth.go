package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"golang.org/x/oauth2"

	//"natsApi/internal/models"
	"github.com/Tutors42Lyon/Mithril/internal/config"
	"github.com/Tutors42Lyon/Mithril/internal/models"
	"github.com/Tutors42Lyon/Mithril/internal/utils"
)

type AuthHandler struct {
	Config   *oauth2.Config
	NatsConn *nats.Conn
}

func NewAuthHandler(nc *nats.Conn, env *config.Env) *AuthHandler {

	return &AuthHandler{
		NatsConn: nc,
		Config: &oauth2.Config{
			ClientID:     env.ClientID,
			ClientSecret: env.ClientSecret,
			RedirectURL:  env.RedirectURL,
			Scopes:       []string{"public"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://api.intra.42.fr/oauth/authorize",
				TokenURL: "https://api.intra.42.fr/oauth/token",
			},
		},
	}
}

func (h *AuthHandler) Login(c *gin.Context) {

	url := h.Config.AuthCodeURL("random") // need to change
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) fetchUserFrom42(ctx context.Context, code string) (*models.User42, error) {

	token, err := h.Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	client := h.Config.Client(ctx, token)
	resp, err := client.Get("https://api.intra.42.fr/v2/me")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var user models.User42
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

func (h *AuthHandler) syncWithWorker(user42 *models.User42) (*models.UserMessage, error) {

	reqMsg := models.UserMessage{
		Username:   user42.Username,
		Email:      user42.Email,
		IntraID:    user42.ID,
		SchoolYear: user42.School_year,
		IsActive:   user42.Is_active,
	}

	reqBytes, err := json.Marshal(reqMsg)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	msg, err := h.NatsConn.Request("user.login", reqBytes, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("nats request failed: %w", err)
	}

	var respMsg models.UserMessage
	if err := json.Unmarshal(msg.Data, &respMsg); err != nil {
		return nil, fmt.Errorf("unmarshal worker response failed: %w", err)
	}

	return &respMsg, nil
}

func (h *AuthHandler) CallBack(c *gin.Context) {

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found"})
		return
	}

	user42, err := h.fetchUserFrom42(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	workerUser, err := h.syncWithWorker(user42)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Login service unavailable"})
		return
	}

	token, err := utils.GenerateJWT(workerUser.Db_id, workerUser.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    workerUser.Username,
		"token":   token,
	})
}
