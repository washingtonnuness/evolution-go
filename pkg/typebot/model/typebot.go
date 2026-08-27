package typebot_model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// TypebotSettings configurações padrão da instância
type TypebotSettings struct {
	ID                string         `gorm:"primaryKey" json:"id"`
	InstanceID        string         `gorm:"index;not null" json:"instanceId"`
	Expire            int            `gorm:"default:300" json:"expire"` // segundos
	KeywordFinish     string         `gorm:"default:#EXIT" json:"keywordFinish"`
	DelayMessage      int            `gorm:"default:1000" json:"delayMessage"` // ms
	UnknownMessage    string         `gorm:"default:Desculpe, não entendi sua mensagem." json:"unknownMessage"`
	ListeningFromMe   bool           `gorm:"default:false" json:"listeningFromMe"`
	StopBotFromMe     bool           `gorm:"default:false" json:"stopBotFromMe"`
	KeepOpen          bool           `gorm:"default:false" json:"keepOpen"`
	DebounceTime      int            `gorm:"default:10" json:"debounceTime"` // segundos
	IgnoreJids        StringArray    `gorm:"type:jsonb" json:"ignoreJids"`
	TypebotIdFallback string         `json:"typebotIdFallback"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// TypebotBot configuração de um bot vinculado à instância
type TypebotBot struct {
	ID              string         `gorm:"primaryKey" json:"id"`
	InstanceID      string         `gorm:"index;not null" json:"instanceId"`
	Enabled         bool           `gorm:"default:true" json:"enabled"`
	Description     string         `json:"description"`
	URL             string         `gorm:"not null" json:"url"`
	Typebot         string         `gorm:"not null" json:"typebot"` // publicId do typebot
	TriggerType     string         `gorm:"default:keyword" json:"triggerType"`    // all, keyword, advanced
	TriggerOperator string         `gorm:"default:equals" json:"triggerOperator"` // equals, contains, startsWith, endsWith, regex
	TriggerValue    string         `json:"triggerValue"`
	Expire          int            `gorm:"default:300" json:"expire"`
	KeywordFinish   string         `gorm:"default:#EXIT" json:"keywordFinish"`
	DelayMessage    int            `gorm:"default:1000" json:"delayMessage"`
	UnknownMessage  string         `json:"unknownMessage"`
	ListeningFromMe bool           `gorm:"default:false" json:"listeningFromMe"`
	StopBotFromMe   bool           `gorm:"default:false" json:"stopBotFromMe"`
	KeepOpen        bool           `gorm:"default:false" json:"keepOpen"`
	DebounceTime    int            `gorm:"default:10" json:"debounceTime"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// IntegrationSession sessão ativa do Typebot
type IntegrationSession struct {
	ID         string         `gorm:"primaryKey" json:"id"`
	RemoteJid  string         `gorm:"index;not null" json:"remoteJid"`
	PushName   string         `json:"pushName"`
	SessionId  string         `gorm:"index;not null" json:"sessionId"` // formato: {random}-{typebotSessionId}
	Status     string         `gorm:"default:opened" json:"status"`    // opened, closed
	AwaitUser  bool           `gorm:"default:false" json:"awaitUser"`
	BotID      string         `gorm:"index" json:"botId"`
	Type       string         `gorm:"default:typebot" json:"type"`
	Parameters JSONMap        `gorm:"type:jsonb" json:"parameters"`
	InstanceID string         `gorm:"index;not null" json:"instanceId"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// JSONMap para campos JSON no PostgreSQL
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// StringArray para arrays de string no PostgreSQL
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

// AutoMigrate roda as migrações
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&TypebotSettings{},
		&TypebotBot{},
		&IntegrationSession{},
	)
}