package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomessguii/logger"

	typebot_dto "github.com/evolution-foundation/evolution-go/pkg/typebot/dto"
	typebot_service "github.com/evolution-foundation/evolution-go/pkg/typebot/service"
)

// TypebotHandler gerencia as requisições HTTP do Typebot
type TypebotHandler struct {
	service typebot_service.TypebotService
}

// NewTypebotHandler cria uma nova instância do handler
func NewTypebotHandler(service typebot_service.TypebotService) TypebotHandler {
	return TypebotHandler{
		service: service,
	}
}

// ==================== SETTINGS ====================

// CreateSettings cria configurações do Typebot
// @Summary Cria configurações do Typebot
// @Tags Typebot
// @Accept json
// @Produce json
// @Param request body typebot_dto.SettingsDTO true "Configurações"
// @Success 200 {object} map[string]interface{}
// @Router /typebot/settings [post]
func (h *TypebotHandler) CreateSettings(c *gin.Context) {
	instanceID := c.GetHeader("instanceId")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instanceId é obrigatório"})
		return
	}

	var dto typebot_dto.SettingsDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service.CreateSettings(instanceID, dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// FindSettings busca configurações do Typebot
// @Summary Busca configurações do Typebot
// @Tags Typebot
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /typebot/settings [get]
func (h *TypebotHandler) FindSettings(c *gin.Context) {
	instanceID := c.GetHeader("instanceId")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instanceId é obrigatório"})
		return
	}

	settings, err := h.service.GetSettings(instanceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateSettings atualiza configurações do Typebot
// @Summary Atualiza configurações do Typebot
// @Tags Typebot
// @Accept json
// @Produce json
// @Param request body typebot_dto.SettingsDTO true "Configurações"
// @Success 200 {object} map[string]interface{}
// @Router /typebot/settings [put]
func (h *TypebotHandler) UpdateSettings(c *gin.Context) {
	instanceID := c.GetHeader("instanceId")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instanceId é obrigatório"})
		return
	}

	var dto typebot_dto.SettingsDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service.UpdateSettings(instanceID, dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// ==================== BOTS ====================

// CreateBot cria um bot do Typebot
// @Summary Cria um bot do Typebot
// @Tags Typebot
// @Accept json
// @Produce json
// @Param request body typebot_dto.BotDTO true "Bot"
// @Success 200 {object} map[string]interface{}
// @Router /typebot/create [post]
func (h *TypebotHandler) CreateBot(c *gin.Context) {
	instanceID := c.GetHeader("instanceId")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instanceId é obrigatório"})
		return
	}

	var dto typebot_dto.BotDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bot, err := h.service.CreateBot(instanceID, dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bot)
}

// FindBots busca todos os bots do Typebot
// @Summary Busca todos os bots do Typebot
// @Tags Typebot
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /typebot/fetch [get]
func (h *TypebotHandler) FindBots(c *gin.Context) {
	instanceID := c.GetHeader("instanceId")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instanceId é obrigatório"})
		return
	}

	bots, err := h.service.FetchBots(instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bots)
}

// UpdateBot atualiza um bot do Typebot
// @Summary Atualiza um bot do Typebot
// @Tags Typebot
// @Accept json
// @Produce json
// @Param botId path string true "ID do bot"
// @Param request body typebot_dto.BotDTO true "Bot"
// @Success 200 {object} map[string]interface{}
// @Router /typebot/update/{botId} [put]
func (h *TypebotHandler) UpdateBot(c *gin.Context) {
	instanceID := c.GetHeader("instanceId")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instanceId é obrigatório"})
		return
	}

	botID := c.Param("botId")

	var dto typebot_dto.BotDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bot, err := h.service.UpdateBot(botID, instanceID, dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bot)
}

// DeleteBot deleta um bot do Typebot
// @Summary Deleta um bot do Typebot
// @Tags Typebot
// @Produce json
// @Param botId path string true "ID do bot"
// @Success 200 {object} map[string]interface{}
// @Router /typebot/delete/{botId} [delete]
func (h *TypebotHandler) DeleteBot(c *gin.Context) {
	instanceID := c.GetHeader("instanceId")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instanceId é obrigatório"})
		return
	}

	botID := c.Param("botId")

	if err := h.service.DeleteBot(botID, instanceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bot deletado com sucesso"})
}

// ==================== SESSIONS ====================

// StartBot inicia um bot do Typebot
// @Summary Inicia um bot do Typebot
// @Tags Typebot
// @Accept json
// @Produce json
// @Param request body typebot_dto.StartBotDTO true "Dados do bot"
// @Success 200 {object} map[string]interface{}
// @Router /typebot/start [post]
func (h *TypebotHandler) StartBot(c *gin.Context) {
	instanceID := c.GetHeader("instanceId")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instanceId é obrigatório"})
		return
	}

	var dto typebot_dto.StartBotDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.StartBot(instanceID, dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ChangeStatus altera o status de uma sessão
// @Summary Altera o status de uma sessão do Typebot
// @Tags Typebot
// @Accept json
// @Produce json
// @Param request body typebot_dto.ChangeStatusDTO true "Status"
// @Success 200 {object} map[string]interface{}
// @Router /typebot/changeStatus [post]
func (h *TypebotHandler) ChangeStatus(c *gin.Context) {
	instanceID := c.GetHeader("instanceId")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instanceId é obrigatório"})
		return
	}

	var dto typebot_dto.ChangeStatusDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ChangeSessionStatus(instanceID, dto); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status alterado com sucesso"})
}

// FetchSessions busca todas as sessões do Typebot
// @Summary Busca todas as sessões do Typebot
// @Tags Typebot
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /typebot/fetchSessions [get]
func (h *TypebotHandler) FetchSessions(c *gin.Context) {
	instanceID := c.GetHeader("instanceId")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instanceId é obrigatório"})
		return
	}

	sessions, err := h.service.FetchSessions(instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// ==================== MÉTODOS AUXILIARES ====================

// Config representa a configuração do Typebot (para uso interno)
type Config struct {
	Enabled         bool              `json:"enabled"`
	URL             string            `json:"url"`
	Typebot         string            `json:"typebot"`
	SessionType     string            `json:"sessionType"`
	Expire          int               `json:"expire"`
	KeywordFinish   string            `json:"keywordFinish"`
	DelayMessage    int               `json:"delayMessage"`
	UnknownMessage  string            `json:"unknownMessage"`
	ListeningFromMe bool              `json:"listeningFromMe"`
	StopBotFromMe   bool              `json:"stopBotFromMe"`
	KeepOpen        bool              `json:"keepOpen"`
	DebounceTime    int               `json:"debounceTime"`
	Variables       map[string]string `json:"variables"`
}

// SendMessageFunc é a função para enviar mensagem de texto
type SendMessageFunc func(ctx context.Context, chatID, text string) error

// HandleMessage processa uma mensagem recebida
func (h *TypebotHandler) HandleMessage(
	ctx context.Context,
	instanceName string,
	config *Config,
	remoteJID string,
	pushName string,
	messageBody string,
	isGroup bool,
	isStatus bool,
	fromMe bool,
	sendText SendMessageFunc,
) error {
	if config == nil || !config.Enabled {
		return nil
	}

	if isStatus || isGroup {
		return nil
	}

	if fromMe && !config.ListeningFromMe {
		return nil
	}

	if config.KeywordFinish != "" && strings.EqualFold(messageBody, config.KeywordFinish) {
		logger.LogInfo("[Typebot] Sessão finalizada por palavra-chave: %s", config.KeywordFinish)
		return nil
	}

	logger.LogDebug("[Typebot] Processando mensagem de %s: %s", remoteJID, messageBody)

	if sendText != nil && messageBody != "" {
		response := fmt.Sprintf("Mensagem recebida: %s", messageBody)
		if err := sendText(ctx, remoteJID, response); err != nil {
			logger.LogError("[Typebot] Erro ao enviar resposta: %v", err)
			return err
		}
	}

	return nil
}

// ValidateConfig valida a configuração
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("configuração não pode ser nula")
	}
	if !config.Enabled {
		return nil
	}
	if config.URL == "" {
		return fmt.Errorf("URL do Typebot é obrigatória")
	}
	if config.Typebot == "" {
		return fmt.Errorf("ID do Typebot é obrigatório")
	}
	if config.SessionType == "" {
		config.SessionType = "preview"
	}
	return nil
}

// BuildStartParams constrói os parâmetros iniciais
func BuildStartParams(instanceName, remoteJID, pushName, messageBody string, config *Config) map[string]interface{} {
	params := make(map[string]interface{})

	params["remoteJid"] = remoteJID
	params["pushName"] = pushName
	params["instanceName"] = instanceName
	params["messageBody"] = messageBody

	for key, value := range config.Variables {
		value = strings.ReplaceAll(value, "{{remoteJid}}", remoteJID)
		value = strings.ReplaceAll(value, "{{pushName}}", pushName)
		value = strings.ReplaceAll(value, "{{instanceName}}", instanceName)
		value = strings.ReplaceAll(value, "{{messageBody}}", messageBody)
		params[key] = value
	}

	return params
}

// Delay calcula o delay entre mensagens
func Delay(ctx context.Context, delayMs int) {
	if delayMs > 0 {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(delayMs) * time.Millisecond):
		}
	}
}