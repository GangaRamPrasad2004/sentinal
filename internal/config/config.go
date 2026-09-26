package config

import (
	"html/template"

	"github.com/GangaRamPrasad2004/sentinal/internal/channeldata"
	"github.com/GangaRamPrasad2004/sentinal/internal/driver"
	"github.com/GangaRamPrasad2004/sentinal/internal/models"
	"github.com/alexedwards/scs/v2"
	"github.com/robfig/cron/v3"
)

// AppConfig holds application configuration
type AppConfig struct {
	DB            *driver.DB
	Session       *scs.SessionManager
	InProduction  bool
	Domain        string
	MonitorMap    map[int]cron.EntryID
	PreferenceMap map[string]string
	Scheduler     *cron.Cron
	WsClient      models.WSClient
	PusherSecret  string
	TemplateCache map[string]*template.Template
	MailQueue     chan channeldata.MailJob
	Version       string
	Identifier    string
}
