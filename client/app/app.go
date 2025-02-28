package app

import (
	"context"
	"crypto/tls"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/beliaevke/simpleKeeper/client/config"
	"github.com/beliaevke/simpleKeeper/client/database"
	"github.com/beliaevke/simpleKeeper/internal/logger"
	"github.com/beliaevke/simpleKeeper/internal/proto"
	"github.com/beliaevke/simpleKeeper/internal/repository/usersrepo"

	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Client struct {
	Cfg          config.ClientFlags
	KeeperClient proto.KeeperClient
	UserLogPass  usersrepo.UserInfo
	UserID       int64
	Token        string
	Key          string
	AESkey       string
	KeyID        int64
	DB           *sql.DB
	App          *tview.Application
	Pages        *tview.Pages
	NotifyCtx    context.Context
	Shutdown     context.CancelFunc
}

func NewClient() *Client {
	return &Client{
		Cfg:   config.NewConfig(),
		App:   tview.NewApplication(),
		Pages: tview.NewPages(),
	}
}

func (Client *Client) Run() error {

	// Инициализация базы данных
	db, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	Client.DB = db
	defer database.CloseDB(Client.DB)

	// Настройка TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	creds := credentials.NewTLS(tlsConfig)

	// Устанавливаем соединение с gRPC сервером
	conn, err := grpc.NewClient(":3200", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// получаем переменную интерфейсного типа KeeperClient,
	// через которую будем отправлять сообщения
	Client.KeeperClient = proto.NewKeeperClient(conn)

	// Создаем канал для сигналов
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	// Создаем контекст, который будет отменен при получении сигнала
	Client.NotifyCtx, Client.Shutdown = signal.NotifyContext(context.Background())
	defer Client.Shutdown()

	go func() {
		<-sigs
		Client.Shutdown() // Отменяем контекст при получении сигнала
	}()

	// Бесконечный цикл для поддержания соединения
	for {
		select {
		case <-Client.NotifyCtx.Done():
			log.Println("Client shutdown...")
			return nil // Завершаем работу клиента
		default:
			ac := &AppClient{Client: Client}

			Client.Pages.AddPage("FormMain", FormMain(ac), true, true)

			if err := Client.App.SetRoot(Client.Pages, true).EnableMouse(true).EnablePaste(true).Run(); err != nil {
				logger.Warnf("Client start fail: " + err.Error())
				return err
			}
			Client.Shutdown()
		}
	}
}
