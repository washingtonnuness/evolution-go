package typebot_dto

import "time"

// SettingsDTO configurações da instância
type SettingsDTO struct {
	Expire           int      `json:"expire"`
	KeywordFinish    string   `json:"keywordFinish"`
	DelayMessage     int      `json:"delayMessage"`
	UnknownMessage   string   `json:"unknownMessage"`
	ListeningFromMe  bool     `json:"listeningFromMe"`
	StopBotFromMe    bool     `json:"stopBotFromMe"`
	KeepOpen         bool     `json:"keepOpen"`
	DebounceTime     int      `json:"debounceTime"`
	IgnoreJids       []string `json:"ignoreJids"`
	TypebotIdFallback string  `json:"typebotIdFallback"`
}

// BotDTO criação/edição de bot
type BotDTO struct {
	Enabled         bool   `json:"enabled"`
	Description     string `json:"description"`
	URL             string `json:"url" binding:"required"`
	Typebot         string `json:"typebot" binding:"required"`
	TriggerType     string `json:"triggerType"`     // all, keyword, advanced
	TriggerOperator string `json:"triggerOperator"` // equals, contains, startsWith, endsWith, regex
	TriggerValue    string `json:"triggerValue"`
	Expire          int    `json:"expire"`
	KeywordFinish   string `json:"keywordFinish"`
	DelayMessage    int    `json:"delayMessage"`
	UnknownMessage  string `json:"unknownMessage"`
	ListeningFromMe bool   `json:"listeningFromMe"`
	StopBotFromMe   bool   `json:"stopBotFromMe"`
	KeepOpen        bool   `json:"keepOpen"`
	DebounceTime    int    `json:"debounceTime"`
}

// StartBotDTO inicia bot manualmente
type StartBotDTO struct {
	RemoteJid    string        `json:"remoteJid" binding:"required"`
	URL          string        `json:"url" binding:"required"`
	Typebot      string        `json:"typebot" binding:"required"`
	StartSession bool          `json:"startSession"`
	Variables    []VariableDTO `json:"variables"`
}

// VariableDTO variáveis para o Typebot
type VariableDTO struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ChangeStatusDTO altera status da sessão
type ChangeStatusDTO struct {
	RemoteJid string `json:"remoteJid" binding:"required"`
	Status    string `json:"status" binding:"required"` // closed, opened, delete
}

// DebounceEntry controle de debounce por usuário
type DebounceEntry struct {
	Message string
	Timer   *time.Timer
}