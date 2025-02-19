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
	app.log.Info().Msg("初始化应用程序...")

	jobs, err := config.ParseConfig(app.configPath)
	if err != nil {
		return fmt.Errorf("配置文件解析失败: %w", err)
	}

	app.log.Info().
		Int("任务数量", len(jobs)).
		Msg("开始调度任务")

	if err := app.initJobs(jobs); err != nil {
		return err
	}

	if app.bot != nil {
		app.initBot()
	}
	app.log.Info().Msg("应用程序初始化完成")
	return nil
}

// initJobs adds and runs each job, logging relevant status.
func (app *App) initJobs(jobs []config.Job) error {
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

		if err := app.manager.RunJob(id); err != nil {
			app.log.Error().
				Err(err).
				Bytes("任务配置", jobConfig).
				Msg("任务运行失败")
			continue
		}

		app.log.Info().
			Bytes("任务配置", jobConfig).
			Str("JobID", id).
			Msg("任务启动成功")
	}
	return nil
}

// initBot starts a Telegram bot handler in a separate goroutine.
func (app *App) initBot() {
	app.log.Info().Msg("启动 Telegram 机器人...")
	go handler.NewBotHandler(app.bot, app.manager).StartDispatching()
}

// cleanup gracefully shuts down running services, saves configuration, and logs termination.
func (app *App) cleanup(sig os.Signal) {
	app.log.Info().Str("Signal", sig.String()).Msg("正在清理资源...")

	if err := config.SaveConfig(app.configPath, app.manager.JobsAllData()); err != nil {
		app.log.Error().Err(err).Msg("保存配置文件失败")
	}

	if app.bot != nil {
		app.bot.MakeRequest("deleteWebhook", tgApi.Params{})
		app.bot.StopReceivingUpdates()
	}

	app.manager.RemoveAll()

	app.log.Info().Msg("所有资源已清理，程序即将退出")
}

// run manages the application lifecycle: initialization, waiting for signals, and cleanup.
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

// main is the entry point of the application.
func main() {
	app := NewApp("./config.yaml")
	if err := app.run(); err != nil {
		log.Panic().Err(err).Msg("应用程序异常退出")
	}
}
