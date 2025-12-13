package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"golang.org/x/oauth2"

	"github.com/Tutors42Lyon/Mithril/internal/config"
	"github.com/Tutors42Lyon/Mithril/internal/models"
	"github.com/Tutors42Lyon/Mithril/internal/utils"
)

type AuthHandler struct {
	Config       *oauth2.Config
	NatsConn     *nats.Conn
	SessionStore sync.Map
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

func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (h *AuthHandler) LoginInit(c *gin.Context) {

	sessionID, err := generateSessionID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session"})
		return
	}

	h.SessionStore.Store(sessionID, "")
	url := h.Config.AuthCodeURL(sessionID)
	c.JSON(http.StatusOK, gin.H{
		"url":        url,
		"session_id": sessionID,
	})
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

	//TODO read msg heck if error http 404 400 500

	var respMsg models.UserMessage
	if err := json.Unmarshal(msg.Data, &respMsg); err != nil {
		return nil, fmt.Errorf("unmarshal worker response failed: %w", err)
	}

	return &respMsg, nil
}

func (h *AuthHandler) CallBack(c *gin.Context) {

	sessionID := c.Query("state")
	code := c.Query("code")

	if code == "" || sessionID == "" {
		c.String(http.StatusBadRequest, "Error: missing code or state!")
		return
	}

	_, ok := h.SessionStore.Load(sessionID)
	if !ok {
		c.String(http.StatusBadRequest, "Error: Session invalide or expire")
	}

	user42, err := h.fetchUserFrom42(c.Request.Context(), code)
	if err != nil {
		c.String(http.StatusInternalServerError, "error: "+err.Error())
		return
	}

	workerUser, err := h.syncWithWorker(user42)
	if err != nil {
		c.String(http.StatusGatewayTimeout, "error: Login service unavailable")
		return
	}

	token, err := utils.GenerateJWT(workerUser.Db_id, workerUser.Role)
	if err != nil {
		c.String(http.StatusInternalServerError, "error: Token generation failed")
		return
	}

	h.SessionStore.Store(sessionID, token)
}

func (h *AuthHandler) PollLogin(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session_id"})
		return
	}

	val, ok := h.SessionStore.Load(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session can't be found"})
		return
	}

	token := val.(string)

	if token == "" {
		c.JSON(http.StatusAccepted, gin.H{"status": "pending"})
		return
	}

	h.SessionStore.Delete(sessionID)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
