package app

import (
	"errors"

	"github.com/beliaevke/simpleKeeper/client/database"
	"github.com/beliaevke/simpleKeeper/client/models"
	"github.com/beliaevke/simpleKeeper/internal/crypt"
	"github.com/beliaevke/simpleKeeper/internal/logger"
	"github.com/beliaevke/simpleKeeper/internal/proto"
	"github.com/beliaevke/simpleKeeper/internal/repository/usersrepo"
	"github.com/beliaevke/simpleKeeper/server/cmd/users"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
)

type ClientInterface interface {
	Register() (bool, string)
	Login() (bool, string)
	LoginLocal() (bool, string)

	AddLogPass(name string, logpass models.LogPass) (bool, string)
	GetSecret(name string) (string, error)
	GetSecretLocal(name string) (string, error)
	DeleteSecret(name string) (bool, error)
	GetSecretsList() (SecretsList, error)
	GetSecretsListLocal() (SecretsList, error)
}

type AppClient struct {
	Client *Client
}

type SecretsList struct {
	bcard   string
	binary  string
	logpass string
	text    string
	names   []string
}

func (ac *AppClient) Login() (bool, string) {
	var errMsg string

	ping, err := ac.Client.KeeperClient.Ping(ac.Client.NotifyCtx, &proto.PingRequest{})
	if err != nil {
		logger.Warnf("Server is not available: " + err.Error())
	}
	if ping == nil || !ping.Available {
		// Сервер недоступен
		return ac.LoginLocal()
	}

	req := proto.LoginRequest{
		Login:    ac.Client.UserLogPass.UserLogin,
		Password: ac.Client.UserLogPass.UserPassword,
	}

	compressor := grpc.UseCompressor(gzip.Name)

	response, err := ac.Client.KeeperClient.Login(ac.Client.NotifyCtx, &req, compressor)
	if err != nil {
		errMsg = "Ошибка аутентификации: " + err.Error()
		return false, errMsg
	} else if response.Status != "" {
		errMsg = "Ошибка входа: " + response.Status
		return false, errMsg
	}

	ac.Client.UserID = response.UserId
	ac.Client.Token = response.Token
	ac.Client.Key = response.Key

	ac.Client.KeyID, ac.Client.AESkey, err = database.GetAESKey(ac.Client.DB, ac.Client.UserID, ac.Client.Cfg.FlagCryptoCert)
	if err != nil {
		errMsg = "Ошибка получения ключа шифрования: " + err.Error()
		return false, errMsg
	}

	ac.Client.UserLogPass = usersrepo.UserInfo{}

	// Проводим первоначальную синхронизацию
	database.SyncData(ac.Client.DB, ac.Client.KeeperClient, ac.Client.NotifyCtx, ac.Client.UserID)

	return true, errMsg
}

func (ac *AppClient) LoginLocal() (bool, string) {
	var errMsg string

	userID, err := database.LoginUser(ac.Client.DB, ac.Client.UserLogPass.UserLogin, ac.Client.UserLogPass.UserPassword)
	if err != nil {
		errMsg = "Ошибка аутентификации: " + err.Error()
		return false, errMsg
	} else if userID == -1 {
		errMsg = "Ошибка входа - пользователь не найден (необходима синхронизация)"
		return false, errMsg
	}

	ac.Client.UserID = int64(userID)

	ac.Client.Token, ac.Client.Key, err = users.AuthenticateUser(userID)
	if err != nil {
		errMsg = "Ошибка аутентификации: " + err.Error()
		return false, errMsg
	}

	ac.Client.KeyID, ac.Client.AESkey, err = database.GetAESKey(ac.Client.DB, ac.Client.UserID, ac.Client.Cfg.FlagCryptoCert)
	if err != nil {
		errMsg = "Ошибка получения ключа шифрования: " + err.Error()
		return false, errMsg
	}

	ac.Client.UserLogPass = usersrepo.UserInfo{}

	// Проводим первоначальную синхронизацию
	database.SyncData(ac.Client.DB, ac.Client.KeeperClient, ac.Client.NotifyCtx, ac.Client.UserID)

	return true, errMsg
}

func (ac *AppClient) Register() (bool, string) {
	var errMsg string

	req := proto.RegisterRequest{
		Login:    ac.Client.UserLogPass.UserLogin,
		Password: ac.Client.UserLogPass.UserPassword,
	}

	compressor := grpc.UseCompressor(gzip.Name)

	response, err := ac.Client.KeeperClient.Register(ac.Client.NotifyCtx, &req, compressor)
	if err != nil {
		errMsg = "Ошибка регистрации: " + err.Error()
		return false, errMsg
	} else if response.Error != "" {
		errMsg = "Ошибка регистрации: " + response.Error
		return false, errMsg
	}

	ac.Client.UserID = response.UserId
	ac.Client.Token = response.Token
	ac.Client.Key = response.Key

	ac.Client.KeyID, ac.Client.AESkey, err = database.GetAESKey(ac.Client.DB, ac.Client.UserID, ac.Client.Cfg.FlagCryptoCert)
	if err != nil {
		errMsg = "Ошибка получения ключа шифрования: " + err.Error()
		return false, errMsg
	}

	ac.Client.UserLogPass = usersrepo.UserInfo{}

	return true, errMsg
}

func (ac *AppClient) AddLogPass(name string, logpass models.LogPass) (bool, string) {
	var errMsg string

	content, err := models.EncodeSecret(logpass)
	if err != nil {
		errMsg = "Ошибка добавления: " + err.Error()
		return false, errMsg
	}

	// Расшифрование AES-ключа с помощью RSA
	keyAES, err := crypt.Decrypt(ac.Client.Cfg.FlagCryptoKey, ac.Client.AESkey)
	if err != nil {
		errMsg = "Error during RSA decryption:" + err.Error()
		return false, errMsg
	}

	ciphertext, err := crypt.EncryptAES(content, keyAES)
	if err != nil {
		errMsg = "Error during encryption:" + err.Error()
		return false, errMsg
	}

	req := proto.CreateSecretRequest{
		Name:    name,
		Type:    "LOGPASS",
		Content: []byte(ciphertext),
		UserId:  ac.Client.UserID,
		KeyId:   ac.Client.KeyID,
	}

	database.AddSecret(ac.Client.DB, &req)

	md := metadata.New(
		map[string]string{
			"X-User-Token": ac.Client.Token,
			"X-User-Key":   ac.Client.Key,
		})

	ctx := metadata.NewOutgoingContext(ac.Client.NotifyCtx, md)

	compressor := grpc.UseCompressor(gzip.Name)

	response, err := ac.Client.KeeperClient.CreateSecret(ctx, &req, compressor)
	if err != nil {
		errMsg = "Ошибка: " + err.Error()
		return false, errMsg
	} else if response == nil {
		errMsg = "Ошибка разбора ответа сервера (CreateSecretResponse - LOGPASS)"
		return false, errMsg
	}

	return true, errMsg
}

func (ac *AppClient) AddText(name string, txt models.Text) (bool, string) {
	var errMsg string

	content, err := models.EncodeSecret(txt)
	if err != nil {
		errMsg = "Ошибка добавления: " + err.Error()
		return false, errMsg
	}

	// Расшифрование AES-ключа с помощью RSA
	keyAES, err := crypt.Decrypt(ac.Client.Cfg.FlagCryptoKey, ac.Client.AESkey)
	if err != nil {
		errMsg = "Error during RSA decryption:" + err.Error()
		return false, errMsg
	}

	ciphertext, err := crypt.EncryptAES(content, keyAES)
	if err != nil {
		errMsg = "Error during encryption:" + err.Error()
		return false, errMsg
	}

	req := proto.CreateSecretRequest{
		Name:    name,
		Type:    "TEXT",
		Content: []byte(ciphertext),
		UserId:  ac.Client.UserID,
		KeyId:   ac.Client.KeyID,
	}

	database.AddSecret(ac.Client.DB, &req)

	md := metadata.New(
		map[string]string{
			"X-User-Token": ac.Client.Token,
			"X-User-Key":   ac.Client.Key,
		})

	ctx := metadata.NewOutgoingContext(ac.Client.NotifyCtx, md)

	compressor := grpc.UseCompressor(gzip.Name)

	response, err := ac.Client.KeeperClient.CreateSecret(ctx, &req, compressor)
	if err != nil {
		errMsg = "Ошибка: " + err.Error()
		return false, errMsg
	} else if response == nil {
		errMsg = "Ошибка разбора ответа сервера (CreateSecretResponse - LOGPASS)"
		return false, errMsg
	}

	return true, errMsg
}

func (ac *AppClient) AddBCard(name string, txt models.BCard) (bool, string) {
	var errMsg string

	content, err := models.EncodeSecret(txt)
	if err != nil {
		errMsg = "Ошибка добавления: " + err.Error()
		return false, errMsg
	}

	// Расшифрование AES-ключа с помощью RSA
	keyAES, err := crypt.Decrypt(ac.Client.Cfg.FlagCryptoKey, ac.Client.AESkey)
	if err != nil {
		errMsg = "Error during RSA decryption:" + err.Error()
		return false, errMsg
	}

	ciphertext, err := crypt.EncryptAES(content, keyAES)
	if err != nil {
		errMsg = "Error during encryption:" + err.Error()
		return false, errMsg
	}

	req := proto.CreateSecretRequest{
		Name:    name,
		Type:    "BCARD",
		Content: []byte(ciphertext),
		UserId:  ac.Client.UserID,
		KeyId:   ac.Client.KeyID,
	}

	database.AddSecret(ac.Client.DB, &req)

	md := metadata.New(
		map[string]string{
			"X-User-Token": ac.Client.Token,
			"X-User-Key":   ac.Client.Key,
		})

	ctx := metadata.NewOutgoingContext(ac.Client.NotifyCtx, md)

	compressor := grpc.UseCompressor(gzip.Name)

	response, err := ac.Client.KeeperClient.CreateSecret(ctx, &req, compressor)
	if err != nil {
		errMsg = "Ошибка: " + err.Error()
		return false, errMsg
	} else if response == nil {
		errMsg = "Ошибка разбора ответа сервера (CreateSecretResponse - BCARD)"
		return false, errMsg
	}

	return true, errMsg
}

func (ac *AppClient) GetSecret(name string) (string, error) {

	ping, err := ac.Client.KeeperClient.Ping(ac.Client.NotifyCtx, &proto.PingRequest{})
	if err != nil {
		logger.Warnf("Server is not available: " + err.Error())
	}
	if ping == nil || !ping.Available {
		// Сервер недоступен
		return ac.GetSecretLocal(name)
	}

	req := proto.GetSecretRequest{
		Name:   name,
		UserId: ac.Client.UserID,
	}
	compressor := grpc.UseCompressor(gzip.Name)
	response, err := ac.Client.KeeperClient.GetSecret(ac.Client.NotifyCtx, &req, compressor)
	if err != nil {
		return "", err
	}

	// Поиск AES-ключа по ID
	AESkey, err := database.GetAESKeyByID(ac.Client.DB, ac.Client.UserID, response.KeyId)
	if err != nil {
		return "", err
	}

	// Расшифрование AES-ключа с помощью RSA
	keyAES, err := crypt.Decrypt(ac.Client.Cfg.FlagCryptoKey, AESkey)
	if err != nil {
		return "", err
	}

	// Расшифровка секрета
	plaintext, err := crypt.DecryptAES(string(response.Content), keyAES)
	if err != nil {
		return "", err
	}

	return plaintext, nil
}

func (ac *AppClient) GetSecretLocal(name string) (string, error) {
	var response *proto.GetSecretResponse

	response, err := database.GetSecret(ac.Client.DB, name, int(ac.Client.UserID))
	if err != nil {
		return "", err
	}

	// Поиск AES-ключа по ID
	AESkey, err := database.GetAESKeyByID(ac.Client.DB, ac.Client.UserID, response.KeyId)
	if err != nil {
		return "", err
	}

	// Расшифрование AES-ключа с помощью RSA
	keyAES, err := crypt.Decrypt(ac.Client.Cfg.FlagCryptoKey, AESkey)
	if err != nil {
		return "", err
	}

	// Расшифровка секрета
	plaintext, err := crypt.DecryptAES(string(response.Content), keyAES)
	if err != nil {
		return "", err
	}

	return plaintext, nil
}

func (ac *AppClient) DeleteSecret(name string) (bool, error) {

	req := proto.DeleteSecretRequest{
		Name:   name,
		UserId: ac.Client.UserID,
	}

	database.DeleteSecret(ac.Client.DB, &req)

	compressor := grpc.UseCompressor(gzip.Name)

	response, err := ac.Client.KeeperClient.DeleteSecret(ac.Client.NotifyCtx, &req, compressor)

	if err != nil {
		logger.Warnf("Client DeleteSecret error: " + err.Error())
		return false, err
	} else if response.Error != "" {
		logger.Warnf("Client DeleteSecret error: " + response.Error)
		return response.Success, errors.New(response.Error)
	}

	return response.Success, nil
}

func (ac *AppClient) GetSecretsList() (SecretsList, error) {
	var secretsList SecretsList

	ping, err := ac.Client.KeeperClient.Ping(ac.Client.NotifyCtx, &proto.PingRequest{})
	if err != nil {
		logger.Warnf("Server is not available: " + err.Error())
	}
	if ping == nil || !ping.Available {
		// Сервер недоступен
		return ac.GetSecretsListLocal()
	}

	req := proto.SecretsListRequest{
		UserId: ac.Client.UserID,
	}

	compressor := grpc.UseCompressor(gzip.Name)

	response, err := ac.Client.KeeperClient.SecretsList(ac.Client.NotifyCtx, &req, compressor)
	if err != nil {
		response.Error = err.Error()
		return secretsList, err
	}

	sep := " | "
	for _, sec := range response.Secrets {
		switch sec.Type {
		case "BCARD":
			secretsList.bcard += sep + sec.Name
		case "BINARY":
			secretsList.binary += sep + sec.Name
		case "LOGPASS":
			secretsList.logpass += sep + sec.Name
		case "TEXT":
			secretsList.text += sep + sec.Name
		default:
			continue
		}
		secretsList.names = append(secretsList.names, sec.Name)
	}

	secretsList.bcard += sep
	secretsList.binary += sep
	secretsList.logpass += sep
	secretsList.text += sep

	return secretsList, nil
}

func (ac *AppClient) GetSecretsListLocal() (SecretsList, error) {
	var secretsList SecretsList
	var secrets []*proto.SecretsList

	secrets, err := database.GetSecretsList(ac.Client.DB, int(ac.Client.UserID))
	if err != nil {
		return secretsList, err
	}

	sep := " | "
	for _, sec := range secrets {
		switch sec.Type {
		case "BCARD":
			secretsList.bcard += sep + sec.Name
		case "BINARY":
			secretsList.binary += sep + sec.Name
		case "LOGPASS":
			secretsList.logpass += sep + sec.Name
		case "TEXT":
			secretsList.text += sep + sec.Name
		default:
			continue
		}
		secretsList.names = append(secretsList.names, sec.Name)
	}

	secretsList.bcard += sep
	secretsList.binary += sep
	secretsList.logpass += sep
	secretsList.text += sep

	return secretsList, nil
}
