package main

import (
	"encoding/json"
	"fmt"
	"github.com/gtoxlili/quantifiable-swap/common/config"
	"github.com/gtoxlili/quantifiable-swap/common/job"
	"github.com/gtoxlili/quantifiable-swap/common/logger"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := logger.NewGeneralLogger()

	if err := runApp(log, "./config.yaml"); err != nil {
		log.Panic().
			Err(err).
			Msg("Application exited with error")
	}
}

func runApp(log zerolog.Logger, configPath string) error {
	log.Info().Msg("应用程序启动中...")
	jobs, err := config.ParseConfig(configPath)
	if err != nil {
		return fmt.Errorf("配置文件解析失败：%w", err)
	}

	manager := job.NewManager()
	log.Info().Int("任务数量", len(jobs)).Msg("开始调度任务")

	for _, j := range jobs {
		jobConfig, _ := json.MarshalIndent(j, "", "  ")
		id, err := manager.AddJob(j)
		if err != nil {
			log.Error().
				Err(err).
				Bytes("任务配置", jobConfig).
				Msg("任务启动失败")
			continue
		}
		_ = manager.RunJob(id)
		log.Info().
			Bytes("任务配置", jobConfig).
			Msg("任务启动成功")
	}

	log.Info().Msg("所有任务已启动，等待终止信号...")
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-signChan

	// 保存任务到配置文件
	if err := config.SaveConfig(configPath, manager.JobsData()); err != nil {
		log.Error().
			Err(err).
			Msg("保存配置文件失败")
	}
	manager.RemoveAll()
	log.Info().
		Str("Sign", sig.String()).
		Msg("接收到终止信号，程序关闭")

	return nil
}
