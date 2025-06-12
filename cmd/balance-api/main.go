package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	balance_api_config "onbloc/internal/config/balance-api"
	handler2 "onbloc/internal/handler"
	"onbloc/internal/repository/postgresdb"
	balance_api_service "onbloc/internal/service/balance-api-service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	path := ""
	flag.StringVar(&path, "c", "config.json", "config path")
	flag.Parse()

	conf, err := balance_api_config.Load(path)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	db, err := gorm.Open(postgres.Open(conf.DB.GetDsn()))
	if err != nil {
		panic(err)
	}

	repository := postgresdb.NewRepository(db)

	service := balance_api_service.NewService(repository)
	handler := handler2.NewBalanceAPIHandler(service)
	r := gin.Default()
	tokenGroup := r.Group("/tokens")
	tokenGroup.GET("/*wildcard", func(c *gin.Context) {
		wildcard := c.Param("wildcard")
		if wildcard == "/balances" {
			handler.GetTokenBalances(c)
		} else if wildcard == "/transfer-history" {
			handler.GetTokenTransferHistory(c)
		} else {
			handler.GetTokenPathBalances(c)
		}
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.Port),
		Handler: r,
	}
	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}
