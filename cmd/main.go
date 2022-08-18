package main

import (
	"context"
	"github.com/FreelancerDenial/todo-app"
	"github.com/FreelancerDenial/todo-app/pkg/handler"
	"github.com/FreelancerDenial/todo-app/pkg/repository"
	"github.com/FreelancerDenial/todo-app/pkg/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	err := initConfig()
	if err != nil {
		logrus.Fatalf("Error while init configs: %s", err.Error())
	}

	err = godotenv.Load()
	if err != nil {
		logrus.Fatalf("Error while loading env: %s", err.Error())
	}

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})
	if err != nil {
		logrus.Fatalf("Error while starting db: %s", err.Error())
	}

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	srv := new(todo.Server)
	go func() {
		err = srv.Run(viper.GetString("port"), handlers.InitRoutes())
		if err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Get error while starting http server: %s", err.Error())
		}
	}()

	logrus.Print("TodoApp Started")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("TodoApp Stopped")

	err = srv.Shutdown(context.Background())
	if err != nil {
		logrus.Errorf("error while server shutdown: %s", err.Error())
	}

	err = db.Close()
	if err != nil {
		logrus.Errorf("error while db conn close: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
