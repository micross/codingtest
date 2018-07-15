package main

import (
	"io"
	"os"

	"github.com/micross/codingtest/models"
	"github.com/micross/codingtest/routes"
	"github.com/micross/codingtest/utils"
	"github.com/micross/codingtest/worker"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func setupConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/codingtest/")
	viper.AddConfigPath("$HOME/.config/codingtest")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	utils.FailOnError(err, "Failed to read config")

	env := viper.GetString("env")
	if env != models.DevelopmentMode {
		gin.SetMode(gin.ReleaseMode)
		// Disable Console Color, you don't need console color when writing the logs to file.
		gin.DisableConsoleColor()
		// Logging to a file.
		fileName := viper.GetString("log_file")
		logFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		utils.FailOnError(err, "Failed to open log file")
		gin.DefaultWriter = io.MultiWriter(logFile)
	}
}

func main() {
	setupConfig()

	models.InitDB()
	models.InitRabbitMQ()
	models.InitRedis()
	go worker.Work()

	// Creates a router without any middleware by default
	app := gin.New()

	// Global middleware
	// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	app.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	app.Use(gin.Recovery())

	routes.Route(app)

	app.Run(viper.GetString("listen"))
}
