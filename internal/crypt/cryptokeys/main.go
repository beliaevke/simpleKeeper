package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/beliaevke/simpleKeeper/client/config"
	"github.com/beliaevke/simpleKeeper/internal/crypt"
	"github.com/beliaevke/simpleKeeper/internal/logger"
	"github.com/beliaevke/simpleKeeper/internal/service"
)

// AESKeyInfo содержит информацию о ключе
type AESKeyInfo struct {
	KeyID string `json:"key_id"`
	Key   []byte `json:"key"`
}

func main() {

	cfg := config.NewConfig()

	// Генерация RSA
	err := crypt.MakeRSACert(
		&crypt.Settings{
			PathToCertificate: cfg.FlagCryptoCert,
			PathToPrivateKey:  cfg.FlagCryptoKey,
		},
	)
	if err != nil {
		logger.Warnf("Error RSA Cert: " + err.Error())
		return
	}

	// Размер ключа AES
	keySize := 32

	// Генерация AES-ключа
	aesKey, err := service.GenerateRandom(keySize)
	if err != nil {
		logger.Warnf("Ошибка генерации ключа:" + err.Error())
		return
	}

	// Генерация уникального идентификатора ключа
	keyID := service.GenerateRandomID()

	// Шифрование AES-ключа с помощью RSA
	encryptedAESKey, err := crypt.Encrypt(cfg.FlagCryptoCert, string(aesKey))
	if err != nil {
		logger.Warnf("Ошибка шифрования AES ключа:" + err.Error())
		return
	}

	// Создание структуры AESKeyInfo
	keyInfo := AESKeyInfo{
		KeyID: keyID,
		Key:   []byte(encryptedAESKey),
	}

	// Сохранение зашифрованного ключа в файл
	data, err := json.Marshal(keyInfo)
	if err != nil {
		logger.Warnf("Ошибка сохранения AES ключа:" + err.Error())
		return
	}
	err = os.WriteFile(cfg.FlagAESKeyInfo, data, 0644)
	if err != nil {
		fmt.Println("Ошибка сохранения AES ключа:", err)
		return
	}

	logger.Infof("Ключ успешно сгенерирован, зашифрован и сохранен в " + cfg.FlagAESKeyInfo)
}
