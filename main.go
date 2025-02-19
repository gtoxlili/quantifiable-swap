package main

import (
	"encoding/json"
	"fmt"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gtoxlili/quantifiable-swap/bot"
	"github.com/gtoxlili/quantifiable-swap/bot/handler"
	"github.com/gtoxlili/quantifiable-swap/common/config"
	"github.com/gtoxlili/quantifiable-swap/common/job"
	"github.com/gtoxlili/quantifiable-swap/common/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	log        zerolog.Logger
	configPath string
	manager    job.IManager
	bot        *tgApi.BotAPI
}

func NewApp(configPath string) *App {
	return &App{
		log:        logger.NewGeneralLogger(),
		configPath: configPath,
		manager:    job.NewManager(),
		bot:        bot.Bot,
	}
}

func (app *App) initialize() error {
	app.log.Info().Msg("应用程序初始化中...")
	jobs, err := config.ParseConfig(app.configPath)
	if err != nil {
		return fmt.Errorf("配置文件解析失败: %w", err)
	}

	app.log.Info().Int("任务数量", len(jobs)).Msg("开始调度任务")
	app.initJobs(jobs)

	if app.bot != nil {
		app.initBot()
	}
	return nil
}

func (app *App) initJobs(jobs []config.Job) {
	for _, j := range jobs {
		jobConfig, _ := json.MarshalIndent(j, "", "  ")
		id, err := app.manager.AddJob(j)
		if err != nil {
			app.log.Error().
				Err(err).
				Bytes("任务配置", jobConfig).
				Msg("任务启动失败")
			continue
		}
		_ = app.manager.RunJob(id)
		app.log.Info().
			Bytes("任务配置", jobConfig).
			Msg("任务启动成功")
	}
}

func (app *App) initBot() {
	app.log.Info().Msg("启动 Telegram 机器人")
	go handler.NewBotHandler(app.bot, app.manager).StartDispatching()
}

func (app *App) cleanup(sig os.Signal) {
	if err := config.SaveConfig(app.configPath, app.manager.JobsData()); err != nil {
		app.log.Error().Err(err).Msg("保存配置文件失败")
	}
	if app.bot != nil {
		app.bot.MakeRequest("deleteWebhook", tgApi.Params{})
		app.bot.StopReceivingUpdates()
	}
	app.manager.RemoveAll()

	log.Info().
		Str("Sign", sig.String()).
		Msg("接收到终止信号，程序关闭")
}

func (app *App) run() error {
	if err := app.initialize(); err != nil {
		return err
	}

	app.log.Info().Msg("所有服务已启动，等待终止信号...")

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-signChan

	app.cleanup(sig)
	return nil
}

func main() {
	app := NewApp("./config.yaml")
	if err := app.run(); err != nil {
		app.log.Panic().Err(err).Msg("应用程序异常退出")
	}
}
