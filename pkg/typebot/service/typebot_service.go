package typebot_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/evolution-foundation/evolution-go/pkg/config"
	instance_model "github.com/evolution-foundation/evolution-go/pkg/instance/model"
	typebot_dto "github.com/evolution-foundation/evolution-go/pkg/typebot/dto"
	typebot_model "github.com/evolution-foundation/evolution-go/pkg/typebot/model"
	"github.com/gomessguii/logger"
	"gorm.io/gorm"
)

// ==================== INTERFACES DE ENVIO ====================

// SendServiceInterface define os métodos de envio necessários.
// Usa structs próprias deste pacote para evitar import cycle com sendMessage/service.
type SendServiceInterface interface {
	SendText(text *TextStruct, instance *instance_model.Instance) error
	SendMediaUrl(media *MediaStruct, instance *instance_model.Instance) error
	SendList(list *ListStruct, instance *instance_model.Instance) error
	SendButton(button *ButtonStruct, instance *instance_model.Instance) error
}

// TextStruct representa mensagem de texto
type TextStruct struct {
	Number string
	Text   string
}

// MediaStruct representa mídia
type MediaStruct struct {
	Number string
	Type   string
	Url    string
}

// ListStruct representa lista
type ListStruct struct {
	Number      string
	Title       string
	Description string
	ButtonText  string
	FooterText  string
	Sections    []SectionStruct
	Delay       int
}

// SectionStruct representa seção de lista
type SectionStruct struct {
	Title string
	Rows  []RowStruct
}

// RowStruct representa linha de lista
type RowStruct struct {
	Title       string
	Description string
	RowId       string
}

// ButtonStruct representa botões
type ButtonStruct struct {
	Number       string
	ThumbnailUrl string
	Title        string
	Description  string
	Footer       string
	Buttons      []ButtonItemStruct
}

// ButtonItemStruct representa item de botão
type ButtonItemStruct struct {
	Type        string
	DisplayText string
	Id          string
	CopyCode    string
	URL         string
	PhoneNumber string
	Currency    string
	Name        string
	KeyType     string
	Key         string
}

// ==================== INTERFACE DO SERVIÇO ====================

// TypebotService interface define todos os métodos
type TypebotService interface {
	// Settings
	CreateSettings(instanceID string, dto typebot_dto.SettingsDTO) (*typebot_model.TypebotSettings, error)
	GetSettings(instanceID string) (*typebot_model.TypebotSettings, error)
	UpdateSettings(instanceID string, dto typebot_dto.SettingsDTO) (*typebot_model.TypebotSettings, error)

	// Bots
	CreateBot(instanceID string, dto typebot_dto.BotDTO) (*typebot_model.TypebotBot, error)
	UpdateBot(botID string, instanceID string, dto typebot_dto.BotDTO) (*typebot_model.TypebotBot, error)
	DeleteBot(botID string, instanceID string) error
	FetchBots(instanceID string) ([]typebot_model.TypebotBot, error)
	FindBotById(botID string) (*typebot_model.TypebotBot, error)

	// Sessions
	ChangeSessionStatus(instanceID string, dto typebot_dto.ChangeStatusDTO) error
	FetchSessions(instanceID string) ([]typebot_model.IntegrationSession, error)

	// Core
	StartBot(instanceID string, dto typebot_dto.StartBotDTO) (map[string]interface{}, error)
	ProcessIncomingMessage(instance *instance_model.Instance, msg *MessageInfo) error

	// SetSendService atualiza o serviço de envio
	SetSendService(sendSvc SendServiceInterface)
}

// MessageInfo contém informações da mensagem recebida
type MessageInfo struct {
	RemoteJid string
	PushName  string
	Content   string
	FromMe    bool
	MessageID string
}

// typebotService implementa TypebotService
type typebotService struct {
	db          *gorm.DB
	config      *config.Config
	sendService SendServiceInterface
	debounceMap map[string]*typebot_dto.DebounceEntry
}

// NewTypebotService cria uma nova instância
func NewTypebotService(db *gorm.DB, cfg *config.Config, sendSvc SendServiceInterface) TypebotService {
	return &typebotService{
		db:          db,
		config:      cfg,
		sendService: sendSvc,
		debounceMap: make(map[string]*typebot_dto.DebounceEntry),
	}
}

// SetSendService atualiza o serviço de envio (para evitar import cycle)
func (s *typebotService) SetSendService(sendSvc SendServiceInterface) {
	s.sendService = sendSvc
}

// isLatestVersion indica se deve usar a API mais recente do Typebot
func (s *typebotService) isLatestVersion() bool {
	return s.config.TypebotApiVersion != "legacy" && s.config.TypebotApiVersion != "v1"
}

func defaultSettings() *typebot_model.TypebotSettings {
	return &typebot_model.TypebotSettings{
		Expire:          300,
		KeywordFinish:   "#EXIT",
		DelayMessage:    1000,
		UnknownMessage:  "Desculpe, não entendi.",
		ListeningFromMe: false,
		StopBotFromMe:   false,
		KeepOpen:        false,
		DebounceTime:    10,
	}
}

// ==================== SETTINGS ====================

func (s *typebotService) CreateSettings(instanceID string, dto typebot_dto.SettingsDTO) (*typebot_model.TypebotSettings, error) {
	settings := &typebot_model.TypebotSettings{
		InstanceID:        instanceID,
		Expire:            dto.Expire,
		KeywordFinish:     dto.KeywordFinish,
		DelayMessage:      dto.DelayMessage,
		UnknownMessage:    dto.UnknownMessage,
		ListeningFromMe:   dto.ListeningFromMe,
		StopBotFromMe:     dto.StopBotFromMe,
		KeepOpen:          dto.KeepOpen,
		DebounceTime:      dto.DebounceTime,
		IgnoreJids:        typebot_model.StringArray(dto.IgnoreJids),
		TypebotIdFallback: dto.TypebotIdFallback,
	}
	if err := s.db.Create(settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *typebotService) GetSettings(instanceID string) (*typebot_model.TypebotSettings, error) {
	var settings typebot_model.TypebotSettings
	err := s.db.Where("instance_id = ?", instanceID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *typebotService) UpdateSettings(instanceID string, dto typebot_dto.SettingsDTO) (*typebot_model.TypebotSettings, error) {
	settings, err := s.GetSettings(instanceID)
	if err != nil {
		return s.CreateSettings(instanceID, dto)
	}

	settings.Expire = dto.Expire
	settings.KeywordFinish = dto.KeywordFinish
	settings.DelayMessage = dto.DelayMessage
	settings.UnknownMessage = dto.UnknownMessage
	settings.ListeningFromMe = dto.ListeningFromMe
	settings.StopBotFromMe = dto.StopBotFromMe
	settings.KeepOpen = dto.KeepOpen
	settings.DebounceTime = dto.DebounceTime
	settings.IgnoreJids = typebot_model.StringArray(dto.IgnoreJids)
	settings.TypebotIdFallback = dto.TypebotIdFallback

	if err := s.db.Save(settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

// ==================== BOTS ====================

func (s *typebotService) CreateBot(instanceID string, dto typebot_dto.BotDTO) (*typebot_model.TypebotBot, error) {
	bot := &typebot_model.TypebotBot{
		InstanceID:      instanceID,
		Enabled:         dto.Enabled,
		Description:     dto.Description,
		URL:             strings.TrimSuffix(dto.URL, "/"),
		Typebot:         dto.Typebot,
		TriggerType:     dto.TriggerType,
		TriggerOperator: dto.TriggerOperator,
		TriggerValue:    dto.TriggerValue,
		Expire:          dto.Expire,
		KeywordFinish:   dto.KeywordFinish,
		DelayMessage:    dto.DelayMessage,
		UnknownMessage:  dto.UnknownMessage,
		ListeningFromMe: dto.ListeningFromMe,
		StopBotFromMe:   dto.StopBotFromMe,
		KeepOpen:        dto.KeepOpen,
		DebounceTime:    dto.DebounceTime,
	}
	if err := s.db.Create(bot).Error; err != nil {
		return nil, err
	}
	return bot, nil
}

func (s *typebotService) UpdateBot(botID string, instanceID string, dto typebot_dto.BotDTO) (*typebot_model.TypebotBot, error) {
	bot, err := s.FindBotById(botID)
	if err != nil {
		return nil, err
	}
	if bot.InstanceID != instanceID {
		return nil, fmt.Errorf("bot does not belong to this instance")
	}

	bot.Enabled = dto.Enabled
	bot.Description = dto.Description
	bot.URL = strings.TrimSuffix(dto.URL, "/")
	bot.Typebot = dto.Typebot
	bot.TriggerType = dto.TriggerType
	bot.TriggerOperator = dto.TriggerOperator
	bot.TriggerValue = dto.TriggerValue
	bot.Expire = dto.Expire
	bot.KeywordFinish = dto.KeywordFinish
	bot.DelayMessage = dto.DelayMessage
	bot.UnknownMessage = dto.UnknownMessage
	bot.ListeningFromMe = dto.ListeningFromMe
	bot.StopBotFromMe = dto.StopBotFromMe
	bot.KeepOpen = dto.KeepOpen
	bot.DebounceTime = dto.DebounceTime

	if err := s.db.Save(bot).Error; err != nil {
		return nil, err
	}
	return bot, nil
}

func (s *typebotService) DeleteBot(botID string, instanceID string) error {
	return s.db.Where("id = ? AND instance_id = ?", botID, instanceID).Delete(&typebot_model.TypebotBot{}).Error
}

func (s *typebotService) FetchBots(instanceID string) ([]typebot_model.TypebotBot, error) {
	var bots []typebot_model.TypebotBot
	err := s.db.Where("instance_id = ?", instanceID).Find(&bots).Error
	return bots, err
}

func (s *typebotService) FindBotById(botID string) (*typebot_model.TypebotBot, error) {
	var bot typebot_model.TypebotBot
	err := s.db.Where("id = ?", botID).First(&bot).Error
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

// ==================== SESSIONS ====================

func (s *typebotService) ChangeSessionStatus(instanceID string, dto typebot_dto.ChangeStatusDTO) error {
	if dto.Status == "closed" {
		return s.db.Model(&typebot_model.IntegrationSession{}).
			Where("instance_id = ? AND remote_jid = ? AND status = ?", instanceID, dto.RemoteJid, "opened").
			Update("status", "closed").Error
	}
	if dto.Status == "delete" {
		return s.db.Where("instance_id = ? AND remote_jid = ?", instanceID, dto.RemoteJid).
			Delete(&typebot_model.IntegrationSession{}).Error
	}
	return nil
}

func (s *typebotService) FetchSessions(instanceID string) ([]typebot_model.IntegrationSession, error) {
	var sessions []typebot_model.IntegrationSession
	err := s.db.Where("instance_id = ?", instanceID).Find(&sessions).Error
	return sessions, err
}

// ==================== CORE LOGIC ====================

func (s *typebotService) StartBot(instanceID string, dto typebot_dto.StartBotDTO) (map[string]interface{}, error) {
	if dto.RemoteJid == "status@broadcast" {
		return nil, fmt.Errorf("invalid remoteJid")
	}

	var instance instance_model.Instance
	if err := s.db.Where("id = ?", instanceID).First(&instance).Error; err != nil {
		return nil, fmt.Errorf("instance not found: %v", err)
	}

	settings, _ := s.GetSettings(instanceID)
	if settings == nil {
		settings = defaultSettings()
	}

	prefilled := make(map[string]interface{})
	for _, v := range dto.Variables {
		prefilled[v.Name] = v.Value
	}
	prefilled["remoteJid"] = dto.RemoteJid
	prefilled["instanceName"] = instance.Name
	prefilled["serverUrl"] = s.config.WebhookUrl
	prefilled["apiKey"] = s.config.GlobalApiKey
	prefilled["ownerJid"] = instance.Jid

	if dto.StartSession {
		bot, err := s.findOrCreateBot(instanceID, dto.URL, dto.Typebot, settings)
		if err != nil {
			return nil, err
		}

		s.db.Where("instance_id = ? AND remote_jid = ?", instanceID, dto.RemoteJid).
			Delete(&typebot_model.IntegrationSession{})

		session, resp, err := s.createNewSession(&instance, bot, dto.RemoteJid, "", prefilled)
		if err != nil {
			return nil, err
		}

		if resp != nil && len(resp.Messages) > 0 {
			s.sendWAMessage(&instance, session, settings, dto.RemoteJid, resp.Messages, resp.Input, resp.ClientSideActions)
		} else {
			typebotSessionID := extractTypebotSessionID(session.SessionId)
			resp2, err := s.continueChat(bot, typebotSessionID, "init")
			if err == nil {
				s.db.Model(session).Updates(map[string]interface{}{
					"status":     "opened",
					"await_user": false,
				})
				s.sendWAMessage(&instance, session, settings, dto.RemoteJid, resp2.Messages, resp2.Input, resp2.ClientSideActions)
			}
		}

		return map[string]interface{}{"session": session}, nil
	}

	id := strconv.Itoa(rand.Intn(10000000000))
	url, reqData := s.buildStartChatPayload(strings.TrimSuffix(dto.URL, "/"), dto.Typebot, prefilled)

	resp, err := s.postToTypebot(url, reqData)
	if err != nil {
		return nil, err
	}

	s.sendWAMessage(&instance, nil, settings, dto.RemoteJid, resp.Messages, resp.Input, resp.ClientSideActions)

	s.emitWebhook(&instance, "typebot.start", map[string]interface{}{
		"remoteJid": dto.RemoteJid,
		"url":       strings.TrimSuffix(dto.URL, "/"),
		"typebot":   dto.Typebot,
		"sessionId": id,
	})

	return map[string]interface{}{
		"typebot": map[string]interface{}{
			"url":                dto.URL,
			"remoteJid":          dto.RemoteJid,
			"typebot":            dto.Typebot,
			"prefilledVariables": prefilled,
		},
		"sessionId": id,
	}, nil
}

func (s *typebotService) ProcessIncomingMessage(instance *instance_model.Instance, msg *MessageInfo) error {
	if msg.RemoteJid == "status@broadcast" {
		return nil
	}

	settings, err := s.GetSettings(instance.Id)
	if err != nil || settings == nil {
		settings = defaultSettings()
	}

	for _, jid := range settings.IgnoreJids {
		if jid == msg.RemoteJid {
			return nil
		}
	}

	if msg.FromMe && !settings.ListeningFromMe {
		return nil
	}

	if msg.FromMe && settings.StopBotFromMe {
		s.db.Where("instance_id = ? AND remote_jid = ?", instance.Id, msg.RemoteJid).
			Delete(&typebot_model.IntegrationSession{})
		return nil
	}

	// Debounce
	if settings.DebounceTime > 0 {
		debounceKey := instance.Id + ":" + msg.RemoteJid
		if entry, exists := s.debounceMap[debounceKey]; exists {
			entry.Message += " " + msg.Content
			if entry.Timer != nil {
				entry.Timer.Stop()
			}
			entry.Timer = time.AfterFunc(time.Duration(settings.DebounceTime)*time.Second, func() {
				delete(s.debounceMap, debounceKey)
				msgCopy := *msg
				msgCopy.Content = entry.Message
				s.processMessageInternal(instance, &msgCopy, settings)
			})
			return nil
		}

		s.debounceMap[debounceKey] = &typebot_dto.DebounceEntry{
			Message: msg.Content,
			Timer: time.AfterFunc(time.Duration(settings.DebounceTime)*time.Second, func() {
				delete(s.debounceMap, debounceKey)
				s.processMessageInternal(instance, msg, settings)
			}),
		}
		return nil
	}

	return s.processMessageInternal(instance, msg, settings)
}

func (s *typebotService) processMessageInternal(instance *instance_model.Instance, msg *MessageInfo, settings *typebot_model.TypebotSettings) error {
	var session typebot_model.IntegrationSession
	sessErr := s.db.Where("instance_id = ? AND remote_jid = ? AND status = ?", instance.Id, msg.RemoteJid, "opened").
		Order("updated_at desc").First(&session).Error

	var bot *typebot_model.TypebotBot
	var botErr error
	if sessErr == nil {
		bot, botErr = s.FindBotById(session.BotID)
	} else {
		bot, botErr = s.findMatchingBot(instance.Id, msg.Content, settings)
	}

	if sessErr != nil && botErr != nil {
		return nil
	}

	// Handle session expiration
	if sessErr == nil && settings.Expire > 0 {
		if time.Since(session.UpdatedAt).Minutes() > float64(settings.Expire) {
			if settings.KeepOpen {
				s.db.Model(&session).Update("status", "closed")
			} else {
				s.db.Delete(&session)
			}

			if bot != nil {
				newSess, resp, err := s.createNewSession(instance, bot, msg.RemoteJid, msg.PushName, nil)
				if err != nil {
					return err
				}
				session = *newSess
				s.sendWAMessage(instance, &session, settings, msg.RemoteJid, resp.Messages, resp.Input, resp.ClientSideActions)
				return nil
			}
		}
	}

	// Create a new session when none exists
	if sessErr != nil && bot != nil {
		newSess, resp, err := s.createNewSession(instance, bot, msg.RemoteJid, msg.PushName, nil)
		if err != nil {
			return err
		}
		session = *newSess

		if resp != nil && len(resp.Messages) > 0 {
			s.sendWAMessage(instance, &session, settings, msg.RemoteJid, resp.Messages, resp.Input, resp.ClientSideActions)
			return nil
		}

		if msg.Content == "" {
			if settings.UnknownMessage != "" {
				s.sendTextMessage(instance, msg.RemoteJid, settings.UnknownMessage, settings.DelayMessage)
			}
			return nil
		}

		if settings.KeywordFinish != "" && strings.EqualFold(msg.Content, settings.KeywordFinish) {
			s.closeSession(instance, &session, settings, msg.RemoteJid)
			return nil
		}

		typebotSessionID := extractTypebotSessionID(session.SessionId)
		resp2, err := s.continueChat(bot, typebotSessionID, msg.Content)
		if err != nil {
			logger.LogError("[Typebot] Error continuing chat: %v", err)
			if settings.UnknownMessage != "" {
				s.sendTextMessage(instance, msg.RemoteJid, settings.UnknownMessage, settings.DelayMessage)
			}
			return nil
		}

		s.db.Model(&session).Updates(map[string]interface{}{
			"status":     "opened",
			"await_user": false,
			"updated_at": time.Now(),
		})
		s.sendWAMessage(instance, &session, settings, msg.RemoteJid, resp2.Messages, resp2.Input, resp2.ClientSideActions)
		return nil
	}

	if sessErr != nil && bot == nil {
		return nil
	}

	// Existing open session
	if session.Status != "opened" {
		return nil
	}

	if settings.KeywordFinish != "" && strings.EqualFold(msg.Content, settings.KeywordFinish) {
		s.closeSession(instance, &session, settings, msg.RemoteJid)
		return nil
	}

	if msg.Content == "" {
		if settings.UnknownMessage != "" {
			s.sendTextMessage(instance, msg.RemoteJid, settings.UnknownMessage, settings.DelayMessage)
		}
		return nil
	}

	typebotSessionID := extractTypebotSessionID(session.SessionId)
	resp, err := s.continueChat(bot, typebotSessionID, msg.Content)
	if err != nil {
		logger.LogError("[Typebot] Error continuing chat: %v", err)
		if settings.UnknownMessage != "" {
			s.sendTextMessage(instance, msg.RemoteJid, settings.UnknownMessage, settings.DelayMessage)
		}
		return nil
	}

	s.db.Model(&session).Updates(map[string]interface{}{
		"status":     "opened",
		"await_user": false,
		"updated_at": time.Now(),
	})

	s.sendWAMessage(instance, &session, settings, msg.RemoteJid, resp.Messages, resp.Input, resp.ClientSideActions)
	return nil
}

func (s *typebotService) closeSession(instance *instance_model.Instance, session *typebot_model.IntegrationSession, settings *typebot_model.TypebotSettings, remoteJid string) {
	statusChange := "delete"
	if settings.KeepOpen {
		statusChange = "closed"
		s.db.Model(session).Update("status", "closed")
	} else {
		s.db.Where("id = ?", session.ID).Delete(&typebot_model.IntegrationSession{})
	}

	s.emitWebhook(instance, "typebot.changeStatus", map[string]interface{}{
		"remoteJid": remoteJid,
		"status":    statusChange,
	})
}

// ==================== TYPEBOT HTTP CLIENT ====================

func (s *typebotService) buildStartChatPayload(url, typebot string, prefilled map[string]interface{}) (string, interface{}) {
	if s.isLatestVersion() {
		reqURL := fmt.Sprintf("%s/api/v1/typebots/%s/startChat", url, typebot)
		reqData := map[string]interface{}{"prefilledVariables": prefilled}
		return reqURL, reqData
	}
	reqURL := fmt.Sprintf("%s/api/v1/sendMessage", url)
	reqData := map[string]interface{}{
		"startParams": map[string]interface{}{
			"publicId":          typebot,
			"prefilledVariables": prefilled,
		},
	}
	return reqURL, reqData
}

func (s *typebotService) continueChat(bot *typebot_model.TypebotBot, typebotSessionID, content string) (*TypebotResponse, error) {
	if s.isLatestVersion() {
		reqURL := fmt.Sprintf("%s/api/v1/sessions/%s/continueChat", bot.URL, typebotSessionID)
		reqData := map[string]interface{}{"message": content}
		return s.postToTypebot(reqURL, reqData)
	}
	reqURL := fmt.Sprintf("%s/api/v1/sendMessage", bot.URL)
	reqData := map[string]interface{}{
		"message":   content,
		"sessionId": typebotSessionID,
	}
	return s.postToTypebot(reqURL, reqData)
}

func (s *typebotService) createNewSession(instance *instance_model.Instance, bot *typebot_model.TypebotBot, remoteJid, pushName string, prefilledVars map[string]interface{}) (*typebot_model.IntegrationSession, *TypebotResponse, error) {
	id := strconv.Itoa(rand.Intn(10000000000))

	if prefilledVars == nil {
		prefilledVars = make(map[string]interface{})
	}
	prefilledVars["remoteJid"] = remoteJid
	if pushName != "" {
		prefilledVars["pushName"] = pushName
	}
	prefilledVars["instanceName"] = instance.Name
	prefilledVars["serverUrl"] = s.config.WebhookUrl
	prefilledVars["apiKey"] = s.config.GlobalApiKey
	prefilledVars["ownerJid"] = instance.Jid

	url, reqData := s.buildStartChatPayload(bot.URL, bot.Typebot, prefilledVars)

	resp, err := s.postToTypebot(url, reqData)
	if err != nil {
		return nil, nil, err
	}

	session := &typebot_model.IntegrationSession{
		RemoteJid:  remoteJid,
		PushName:   pushName,
		SessionId:  fmt.Sprintf("%s-%s", id, resp.SessionID),
		Status:     "opened",
		AwaitUser:  false,
		BotID:      bot.ID,
		Type:       "typebot",
		Parameters: typebot_model.JSONMap(prefilledVars),
		InstanceID: instance.Id,
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, nil, err
	}

	s.emitWebhook(instance, "typebot.start", map[string]interface{}{
		"remoteJid": remoteJid,
		"url":       bot.URL,
		"typebot":   bot.Typebot,
		"sessionId": session.SessionId,
	})

	return session, resp, nil
}

func extractTypebotSessionID(sessionID string) string {
	parts := strings.SplitN(sessionID, "-", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return sessionID
}

func (s *typebotService) sendWAMessage(instance *instance_model.Instance, session *typebot_model.IntegrationSession, settings *typebot_model.TypebotSettings, remoteJid string, messages []TypebotMessage, input *TypebotInput, clientSideActions []ClientSideAction) {
	for _, msg := range messages {
		switch msg.Type {
		case "text":
			text := s.formatRichText(msg.Content.RichText)
			if strings.Contains(text, "[list]") {
				s.processListMessage(instance, text, remoteJid)
			} else if strings.Contains(text, "[buttons]") {
				s.processButtonMessage(instance, text, remoteJid)
			} else {
				s.sendTextMessage(instance, remoteJid, text, settings.DelayMessage)
			}
		case "image":
			if msg.Content.URL != "" {
				s.sendMediaMessage(instance, remoteJid, "image", msg.Content.URL, settings.DelayMessage)
			}
		case "video":
			if msg.Content.URL != "" {
				s.sendMediaMessage(instance, remoteJid, "video", msg.Content.URL, settings.DelayMessage)
			}
		case "audio":
			if msg.Content.URL != "" {
				s.sendMediaMessage(instance, remoteJid, "audio", msg.Content.URL, settings.DelayMessage)
			}
		}

		for _, action := range clientSideActions {
			if action.LastBubbleBlockId == msg.ID && action.Wait != nil {
				time.Sleep(time.Duration(action.Wait.SecondsToWaitFor) * time.Second)
			}
		}
	}

	if input != nil {
		if input.Type == "choice input" {
			var opts []string
			for _, item := range input.Items {
				opts = append(opts, "▶️ "+item.Content)
			}
			s.sendTextMessage(instance, remoteJid, strings.Join(opts, "\n"), settings.DelayMessage)
		}
		if session != nil {
			s.db.Model(session).Update("await_user", true)
		}
	} else {
		if session != nil {
			if settings.KeepOpen {
				s.db.Model(session).Update("status", "closed")
			} else {
				s.db.Delete(session)
			}
		}
		status := "closed"
		if session == nil {
			status = "delete"
		}
		s.emitWebhook(instance, "typebot.changeStatus", map[string]interface{}{
			"remoteJid": remoteJid,
			"status":    status,
		})
	}
}

// ==================== FORMATTING & SENDING ====================

func (s *typebotService) formatRichText(richText []RichTextElement) string {
	var result string
	for _, rt := range richText {
		for _, child := range rt.Children {
			result += s.applyFormatting(child)
		}
		result += "\n"
	}
	result = strings.ReplaceAll(result, "**", "")
	result = strings.ReplaceAll(result, "__", "")
	result = strings.ReplaceAll(result, "~~", "")
	return strings.TrimRight(result, "\n")
}

func (s *typebotService) applyFormatting(el Element) string {
	text := el.Text
	if text == "" && len(el.Children) > 0 && el.Type != "a" {
		for _, child := range el.Children {
			text += s.applyFormatting(child)
		}
	}

	if el.Type == "p" {
		text = strings.TrimSpace(text) + "\n"
	}
	if el.Type == "inline-variable" {
		text = strings.TrimSpace(text)
	}
	if el.Type == "ol" {
		lines := strings.Split(text, "\n")
		var numbered []string
		for i, line := range lines {
			if strings.TrimSpace(line) != "" {
				numbered = append(numbered, fmt.Sprintf("%d. %s", i+1, line))
			}
		}
		text = "\n" + strings.Join(numbered, "\n")
	}
	if el.Type == "li" {
		lines := strings.Split(text, "\n")
		var indented []string
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				indented = append(indented, "  "+line)
			}
		}
		text = strings.Join(indented, "\n")
	}

	formats := ""
	if el.Bold {
		formats += "*"
	}
	if el.Italic {
		formats += "_"
	}
	if el.Underline {
		formats += "~"
	}

	formatted := fmt.Sprintf("%s%s%s", formats, text, reverseString(formats))

	if el.URL != "" {
		if len(el.Children) > 0 && el.Children[0].Text != "" {
			formatted = fmt.Sprintf("[%s]\n(%s)", formatted, el.URL)
		} else {
			formatted = el.URL
		}
	}

	return formatted
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func (s *typebotService) processListMessage(instance *instance_model.Instance, text, remoteJid string) {
	listData := parseListMarkup(text)
	if listData.Title == "" && len(listData.Sections) == 0 {
		s.sendTextMessage(instance, remoteJid, text, 0)
		return
	}

	if err := s.sendService.SendList(&listData, instance); err != nil {
		logger.LogError("[Typebot] Error sending list: %v", err)
	}
}

func (s *typebotService) processButtonMessage(instance *instance_model.Instance, text, remoteJid string) {
	buttonData := parseButtonMarkup(text)
	if buttonData.Title == "" && len(buttonData.Buttons) == 0 {
		s.sendTextMessage(instance, remoteJid, text, 0)
		return
	}

	if err := s.sendService.SendButton(&buttonData, instance); err != nil {
		logger.LogError("[Typebot] Error sending button: %v", err)
	}
}

func (s *typebotService) sendTextMessage(instance *instance_model.Instance, number, text string, delay int) {
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	if err := s.sendService.SendText(&TextStruct{Number: number, Text: text}, instance); err != nil {
		logger.LogError("[Typebot] Error sending text: %v", err)
	}
}

func (s *typebotService) sendMediaMessage(instance *instance_model.Instance, number, mediaType, url string, delay int) {
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	if err := s.sendService.SendMediaUrl(&MediaStruct{Number: number, Type: mediaType, Url: url}, instance); err != nil {
		logger.LogError("[Typebot] Error sending media: %v", err)
	}
}

// ==================== MATCHING LOGIC ====================

func (s *typebotService) findMatchingBot(instanceID, content string, settings *typebot_model.TypebotSettings) (*typebot_model.TypebotBot, error) {
	bots, err := s.FetchBots(instanceID)
	if err != nil {
		return nil, err
	}

	for i := range bots {
		bot := bots[i]
		if !bot.Enabled {
			continue
		}
		if s.matchesTrigger(&bot, content) {
			return &bot, nil
		}
	}

	if settings.TypebotIdFallback != "" {
		var fallback typebot_model.TypebotBot
		err := s.db.Where("id = ? AND instance_id = ?", settings.TypebotIdFallback, instanceID).First(&fallback).Error
		if err == nil {
			return &fallback, nil
		}
	}

	return nil, fmt.Errorf("no matching bot found")
}

func (s *typebotService) matchesTrigger(bot *typebot_model.TypebotBot, content string) bool {
	switch bot.TriggerType {
	case "all":
		return true
	case "keyword":
		switch bot.TriggerOperator {
		case "equals":
			return strings.EqualFold(content, bot.TriggerValue)
		case "contains":
			return strings.Contains(strings.ToLower(content), strings.ToLower(bot.TriggerValue))
		case "startsWith":
			return strings.HasPrefix(strings.ToLower(content), strings.ToLower(bot.TriggerValue))
		case "endsWith":
			return strings.HasSuffix(strings.ToLower(content), strings.ToLower(bot.TriggerValue))
		case "regex":
			matched, _ := regexp.MatchString(bot.TriggerValue, content)
			return matched
		}
	case "advanced":
		matched, _ := regexp.MatchString(bot.TriggerValue, content)
		return matched
	}
	return false
}

func (s *typebotService) findOrCreateBot(instanceID, url, typebot string, settings *typebot_model.TypebotSettings) (*typebot_model.TypebotBot, error) {
	url = strings.TrimSuffix(url, "/")

	var bot typebot_model.TypebotBot
	err := s.db.Where("instance_id = ? AND url = ? AND typebot = ?", instanceID, url, typebot).First(&bot).Error
	if err == nil {
		return &bot, nil
	}

	bot = typebot_model.TypebotBot{
		InstanceID:      instanceID,
		Enabled:         true,
		URL:             url,
		Typebot:         typebot,
		TriggerType:     "all",
		Expire:          settings.Expire,
		KeywordFinish:   settings.KeywordFinish,
		DelayMessage:    settings.DelayMessage,
		UnknownMessage:  settings.UnknownMessage,
		ListeningFromMe: settings.ListeningFromMe,
		StopBotFromMe:   settings.StopBotFromMe,
		KeepOpen:        settings.KeepOpen,
	}
	if err := s.db.Create(&bot).Error; err != nil {
		return nil, err
	}
	return &bot, nil
}

// ==================== HTTP CLIENT ====================

type TypebotResponse struct {
	SessionID         string             `json:"sessionId"`
	Messages          []TypebotMessage   `json:"messages"`
	Input             *TypebotInput      `json:"input"`
	ClientSideActions []ClientSideAction `json:"clientSideActions"`
}

type TypebotMessage struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"`
	Content MessageContent `json:"content"`
}

type MessageContent struct {
	RichText []RichTextElement `json:"richText"`
	URL      string            `json:"url"`
	Alt      string            `json:"alt,omitempty"`
}

type RichTextElement struct {
	Children []Element `json:"children"`
}

type Element struct {
	Type      string    `json:"type"`
	Text      string    `json:"text"`
	Children  []Element `json:"children"`
	Bold      bool      `json:"bold"`
	Italic    bool      `json:"italic"`
	Underline bool      `json:"underline"`
	URL       string    `json:"url"`
}

type TypebotInput struct {
	Type  string      `json:"type"`
	Items []InputItem `json:"items"`
}

type InputItem struct {
	Content string `json:"content"`
}

type ClientSideAction struct {
	LastBubbleBlockId string `json:"lastBubbleBlockId"`
	Wait              *struct {
		SecondsToWaitFor int `json:"secondsToWaitFor"`
	} `json:"wait"`
}

func (s *typebotService) postToTypebot(url string, data interface{}) (*TypebotResponse, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("typebot returned status %d", resp.StatusCode)
	}

	var result TypebotResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ==================== WEBHOOK / EVENTS ====================

func (s *typebotService) emitWebhook(instance *instance_model.Instance, event string, data map[string]interface{}) {
	payload := map[string]interface{}{
		"event":     event,
		"instance":  instance.Name,
		"data":      data,
		"timestamp": time.Now(),
	}
	jsonData, _ := json.Marshal(payload)
	logger.LogDebug("[Typebot] Webhook: %s", string(jsonData))
}

// ==================== MARKUP PARSERS ====================

func parseListMarkup(text string) ListStruct {
	data := ListStruct{ButtonText: "Clique aqui"}

	if match := extractBetween(text, "[title]", "[/title]"); match != "" {
		data.Title = match
	}
	if match := extractBetween(text, "[description]", "[/description]"); match != "" {
		data.Description = match
	}
	if match := extractBetween(text, "[buttonText]", "[/buttonText]"); match != "" {
		data.ButtonText = match
	}
	if match := extractBetween(text, "[footerText]", "[/footerText]"); match != "" {
		data.FooterText = match
	}

	sectionRegex := regexp.MustCompile(`(?s)\[section\](.*?)\[/section\]`)
	sectionMatches := sectionRegex.FindAllStringSubmatch(text, -1)

	for _, sectionMatch := range sectionMatches {
		section := SectionStruct{
			Title: extractBetween(sectionMatch[1], "[title]", "[/title]"),
		}

		rowRegex := regexp.MustCompile(`(?s)\[row\](.*?)\[/row\]`)
		rowMatches := rowRegex.FindAllStringSubmatch(sectionMatch[1], -1)

		for _, rowMatch := range rowMatches {
			row := RowStruct{
				Title:       extractBetween(rowMatch[1], "[title]", "[/title]"),
				Description: extractBetween(rowMatch[1], "[description]", "[/description]"),
				RowId:       extractBetween(rowMatch[1], "[rowId]", "[/rowId]"),
			}
			section.Rows = append(section.Rows, row)
		}

		data.Sections = append(data.Sections, section)
	}

	return data
}

func parseButtonMarkup(text string) ButtonStruct {
	data := ButtonStruct{}

	if match := extractBetween(text, "[thumbnailUrl]", "[/thumbnailUrl]"); match != "" {
		data.ThumbnailUrl = match
	}
	if match := extractBetween(text, "[title]", "[/title]"); match != "" {
		data.Title = match
	}
	if match := extractBetween(text, "[description]", "[/description]"); match != "" {
		data.Description = match
	}
	if match := extractBetween(text, "[footer]", "[/footer]"); match != "" {
		data.Footer = match
	}

	buttonRegex := regexp.MustCompile(`(?s)\[(reply|pix|copy|call|url)\](.*?)\[/\1\]`)
	buttonMatches := buttonRegex.FindAllStringSubmatch(text, -1)

	for _, btnMatch := range buttonMatches {
		btnType := btnMatch[1]
		btnContent := btnMatch[2]

		button := ButtonItemStruct{Type: btnType}

		switch btnType {
		case "pix":
			button.Currency = extractBetween(btnContent, "[currency]", "[/currency]")
			button.Name = extractBetween(btnContent, "[name]", "[/name]")
			button.KeyType = extractBetween(btnContent, "[keyType]", "[/keyType]")
			button.Key = extractBetween(btnContent, "[key]", "[/key]")
		case "reply":
			button.DisplayText = extractBetween(btnContent, "[displayText]", "[/displayText]")
			button.Id = extractBetween(btnContent, "[id]", "[/id]")
		case "copy":
			button.DisplayText = extractBetween(btnContent, "[displayText]", "[/displayText]")
			button.CopyCode = extractBetween(btnContent, "[copyCode]", "[/copyCode]")
		case "call":
			button.DisplayText = extractBetween(btnContent, "[displayText]", "[/displayText]")
			button.PhoneNumber = extractBetween(btnContent, "[phone]", "[/phone]")
		case "url":
			button.DisplayText = extractBetween(btnContent, "[displayText]", "[/displayText]")
			button.URL = extractBetween(btnContent, "[url]", "[/url]")
		}

		if button.DisplayText != "" || button.Key != "" || button.URL != "" || button.PhoneNumber != "" || button.CopyCode != "" {
			data.Buttons = append(data.Buttons, button)
		}
	}

	return data
}

func extractBetween(text, start, end string) string {
	startIdx := strings.Index(text, start)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(start)
	endIdx := strings.Index(text[startIdx:], end)
	if endIdx == -1 {
		return ""
	}
	return strings.TrimSpace(text[startIdx : startIdx+endIdx])
}
