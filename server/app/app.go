package app

import (
	"context"
	"crypto/tls"
	"os"
	"os/signal"
	"syscall"

	"github.com/beliaevke/simpleKeeper/internal/db/migrations"
	"github.com/beliaevke/simpleKeeper/internal/db/postgres"
	"github.com/beliaevke/simpleKeeper/internal/logger"
	"github.com/beliaevke/simpleKeeper/internal/proto"
	"github.com/beliaevke/simpleKeeper/internal/repository/secretsrepo"
	"github.com/beliaevke/simpleKeeper/internal/repository/syncrepo"
	"github.com/beliaevke/simpleKeeper/internal/repository/usersrepo"
	"github.com/beliaevke/simpleKeeper/internal/service"
	"github.com/beliaevke/simpleKeeper/server/cmd/secrets"
	"github.com/beliaevke/simpleKeeper/server/cmd/syncs"
	"github.com/beliaevke/simpleKeeper/server/cmd/users"
	"github.com/beliaevke/simpleKeeper/server/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type srv struct {
	// implement GRPC server
	proto.UnimplementedKeeperServer
	// GRPC server
	gRPCServer *grpc.Server
	// TLS
	tlsd *service.TLSData
	// DB
	db *postgres.DB
	// Config
	cfg config.ServerFlags
}

func NewServer() *srv {
	cfg := config.NewConfig()
	return &srv{
		cfg:  cfg,
		tlsd: service.NewTLSData(cfg.FlagTLSCert, cfg.EnvTLSKey),
	}
}

func (srv *srv) Run() error {

	ctx := context.Background()

	db, err := postgres.NewDB(ctx, srv.cfg)
	if err != nil {
		logger.Warnf("SetDB fail: " + err.Error())
		return err
	}

	if err := migrations.Run(srv.cfg, ctx); err != nil {
		return err
	}

	srv.db = db

	// Настройка TLS
	cert, err := tls.LoadX509KeyPair(srv.tlsd.TLSCertPath, srv.tlsd.TLSKeyPath)
	if err != nil {
		logger.Warnf("Failed to load server certificate: \n" + err.Error())
		return err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listener, err := tls.Listen("tcp", ":3200", tlsConfig)
	if err != nil {
		logger.Warnf("gRPC Server error: " + err.Error())
		return err
	}
	defer listener.Close()

	srv.gRPCServer = grpc.NewServer(grpc.ChainUnaryInterceptor(srv.tknInterceptor))

	proto.RegisterKeeperServer(srv.gRPCServer, srv)
	logger.Infof("gRPC server started...")

	go func() {
		if err := srv.gRPCServer.Serve(listener); err != nil {
			logger.Warnf("gRPC listen error: " + err.Error())
		}
	}()

	// через этот канал сообщим основному потоку, что соединения закрыты
	idleConnsClosed := make(chan struct{})

	// канал для перенаправления прерываний
	// поскольку нужно отловить всего одно прерывание,
	// ёмкости 1 для канала будет достаточно
	sigs := make(chan os.Signal, 1)

	// регистрируем перенаправление прерываний
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	// запускаем горутину обработки пойманных прерываний
	go func() {
		// читаем из канала прерываний
		// поскольку нужно прочитать только одно прерывание,
		// можно обойтись без цикла
		<-sigs
		// получили сигнал os.Interrupt, запускаем процедуру graceful shutdown
		srv.gRPCServer.GracefulStop()
		logger.Infof("gRPC server shutdown...")
		// сообщаем основному потоку,
		// что все сетевые соединения обработаны и закрыты
		close(idleConnsClosed)
	}()

	// ждём завершения процедуры graceful shutdown
	<-idleConnsClosed
	// получили оповещение о завершении
	// здесь можно освобождать ресурсы перед выходом,
	// например закрыть соединение с базой данных,
	// закрыть открытые файлы
	logger.Infof("gRPC server shutdown gracefully")

	return nil

}

func (srv *srv) Ping(ctx context.Context, in *proto.PingRequest) (*proto.PingResponse, error) {
	return &proto.PingResponse{Available: true}, nil
}

func (srv *srv) tknInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var userToken, userKey string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		paramToken := md.Get("X-User-Token")
		if len(paramToken) > 0 {
			userToken = paramToken[0]
		}
		paramKey := md.Get("X-User-Key")
		if len(paramKey) > 0 {
			userKey = paramKey[0]
		}
	}
	if len(userToken) > 0 {
		isValid, err := users.ValidateToken(userToken, userKey)
		if err != nil {
			return nil, status.Error(codes.Aborted, err.Error())
		} else if !isValid {
			return nil, status.Error(codes.Aborted, "Token is invalid or has expired")
		}
	}

	return handler(ctx, req)
}

func (srv *srv) SyncDataKeys(ctx context.Context, in *proto.SyncDataKeysRequest) (*proto.SyncDataKeysResponse, error) {
	var response proto.SyncDataKeysResponse

	sync := syncrepo.SyncInfo{
		OwnerID: in.OwnerID,
	}

	syncrepo := syncs.NewRepo(srv.db)

	ctx, cancel := context.WithTimeout(ctx, syncrepo.Timeout())
	defer cancel()

	keys, err := syncrepo.SyncDataKeys(ctx, sync)
	if err != nil {
		return &response, err
	}

	response.Keys = keys

	return &response, nil
}

func (srv *srv) PushDataKeys(ctx context.Context, in *proto.PushDataKeysRequest) (*proto.PushDataKeysResponse, error) {
	var response proto.PushDataKeysResponse

	syncrepo := syncs.NewRepo(srv.db)

	ctx, cancel := context.WithTimeout(ctx, syncrepo.Timeout())
	defer cancel()

	err := syncrepo.PushDataKeys(ctx, in.Keys)
	if err != nil {
		return &response, err
	}

	response.Message = "Keys data pushed successfully"

	return &response, nil
}

func (srv *srv) SyncDataUsers(ctx context.Context, in *proto.SyncDataUserRequest) (*proto.SyncDataUserResponse, error) {
	var response proto.SyncDataUserResponse

	sync := syncrepo.SyncInfo{
		OwnerID: in.OwnerID,
	}

	syncrepo := syncs.NewRepo(srv.db)

	ctx, cancel := context.WithTimeout(ctx, syncrepo.Timeout())
	defer cancel()

	users, err := syncrepo.SyncDataUsers(ctx, sync)
	if err != nil {
		return &response, err
	}

	response.Users = users

	return &response, nil
}

func (srv *srv) SyncDataSecrets(ctx context.Context, in *proto.SyncDataSecretsRequest) (*proto.SyncDataSecretsResponse, error) {
	var response proto.SyncDataSecretsResponse

	sync := syncrepo.SyncInfo{
		OwnerID: in.OwnerID,
	}

	syncrepo := syncs.NewRepo(srv.db)

	ctx, cancel := context.WithTimeout(ctx, syncrepo.Timeout())
	defer cancel()

	secrets, err := syncrepo.SyncDataSecrets(ctx, sync)
	if err != nil {
		return &response, err
	}

	response.Secrets = secrets

	return &response, nil
}

func (srv *srv) PushDataSecrets(ctx context.Context, in *proto.PushDataSecretsRequest) (*proto.PushDataSecretsResponse, error) {
	var response proto.PushDataSecretsResponse

	syncrepo := syncs.NewRepo(srv.db)

	ctx, cancel := context.WithTimeout(ctx, syncrepo.Timeout())
	defer cancel()

	err := syncrepo.PushDataSecrets(ctx, in.Secrets)
	if err != nil {
		return &response, err
	}

	response.Message = "Secrets data pushed successfully"

	return &response, nil
}

func (srv *srv) Login(ctx context.Context, in *proto.LoginRequest) (*proto.LoginResponse, error) {
	var response proto.LoginResponse

	user := usersrepo.UserInfo{
		UserLogin:    in.Login,
		UserPassword: in.Password,
	}

	usersrepo := users.NewRepo(srv.db)

	ctx, cancel := context.WithTimeout(ctx, usersrepo.Timeout())
	defer cancel()

	userID, err := usersrepo.LoginUser(ctx, user)
	if err != nil {
		response.Status = err.Error()
		return &response, nil
	}
	if userID == -1 {
		response.Status = "incorrect login or password"
		return &response, nil
	}
	response.Token, response.Key, err = users.AuthenticateUser(userID)
	if err != nil {
		response.Status = err.Error()
		return &response, err
	}

	response.UserId = int64(userID)

	return &response, nil
}

func (srv *srv) Register(ctx context.Context, in *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	var response proto.RegisterResponse

	user := usersrepo.UserInfo{
		UserLogin:    in.Login,
		UserPassword: in.Password,
	}

	usersrepo := users.NewRepo(srv.db)

	ctx, cancel := context.WithTimeout(ctx, usersrepo.Timeout())
	defer cancel()

	userID, err := usersrepo.GetUser(ctx, user)
	if err != nil {
		response.Error = err.Error()
		return &response, nil
	}
	if userID != -1 {
		response.Error = "user already exists with this login"
		return &response, nil
	}
	userID, err = usersrepo.CreateUser(ctx, user)
	if err != nil {
		response.Error = err.Error()
		return &response, nil
	}
	response.Token, response.Key, err = users.AuthenticateUser(userID)
	if err != nil {
		response.Error = err.Error()
		return &response, err
	}

	response.UserId = int64(userID)

	return &response, nil
}

func (srv *srv) CreateSecret(ctx context.Context, in *proto.CreateSecretRequest) (*proto.CreateSecretResponse, error) {
	var response proto.CreateSecretResponse

	secret := secretsrepo.SecretInfo{
		SecretName: in.Name,
		SecretType: in.Type,
		Content:    in.Content,
		OwnerID:    in.UserId,
		KeyID:      in.KeyId,
	}

	secretsrepo := secrets.NewRepo(srv.db)

	ctx, cancel := context.WithTimeout(ctx, secretsrepo.Timeout())
	defer cancel()

	secretID, err := secretsrepo.GetSecret(ctx, secret)
	if err != nil {
		response.Error = err.Error()
		return &response, nil
	}
	if secretID != -1 {
		response.Error = "secret already exists with this name"
		return &response, nil
	}
	secretID, err = secretsrepo.CreateSecret(ctx, secret)
	if err != nil {
		response.Error = err.Error()
		return &response, nil
	}

	return &response, nil
}

func (srv *srv) GetSecret(ctx context.Context, in *proto.GetSecretRequest) (*proto.GetSecretResponse, error) {
	var response *proto.GetSecretResponse

	secret := secretsrepo.SecretInfo{
		SecretName: in.Name,
		SecretType: in.Type,
		OwnerID:    in.UserId,
	}

	secretsrepo := secrets.NewRepo(srv.db)

	ctx, cancel := context.WithTimeout(ctx, secretsrepo.Timeout())
	defer cancel()

	response, err := secretsrepo.GetSecretInfo(ctx, secret)
	if err != nil {
		response.Error = err.Error()
		return response, nil
	}

	return response, nil
}

func (srv *srv) DeleteSecret(ctx context.Context, in *proto.DeleteSecretRequest) (*proto.DeleteSecretResponse, error) {
	var response *proto.DeleteSecretResponse

	secret := secretsrepo.SecretInfo{
		SecretName: in.Name,
		SecretType: in.Type,
		OwnerID:    in.UserId,
	}

	secretsrepo := secrets.NewRepo(srv.db)

	ctx, cancel := context.WithTimeout(ctx, secretsrepo.Timeout())
	defer cancel()

	response, err := secretsrepo.DeleteSecret(ctx, secret)
	if err != nil {
		response.Error = err.Error()
		return response, nil
	}

	return response, nil
}

func (srv *srv) SecretsList(ctx context.Context, in *proto.SecretsListRequest) (*proto.SecretsListResponse, error) {
	var response proto.SecretsListResponse
	var err error

	secretsrepo := secrets.NewRepo(srv.db)

	ctx, cancel := context.WithTimeout(ctx, secretsrepo.Timeout())
	defer cancel()

	response.Secrets, err = secretsrepo.GetSecretsList(ctx, in.UserId)
	if err != nil {
		response.Error = err.Error()
		return &response, nil
	}

	return &response, nil
}
