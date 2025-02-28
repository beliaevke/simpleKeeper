package database

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"

	"github.com/beliaevke/simpleKeeper/internal/crypt"
	"github.com/beliaevke/simpleKeeper/internal/logger"
	"github.com/beliaevke/simpleKeeper/internal/proto"
	"github.com/beliaevke/simpleKeeper/internal/service"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	_ "github.com/mattn/go-sqlite3" // Импортируем драйвер SQLite
)

// InitDB инициализирует базу данных и создает таблицы
func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./database/data.db")
	if err != nil {
		return nil, err
	}

	// Применение миграций
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migrations",
		"sqlite3://database/migrations/data.db", driver)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	return db, nil
}

// AddSecret добавляет новую запись в базу данных
func AddSecret(db *sql.DB, s *proto.CreateSecretRequest) error {
	insertSQL :=
		`INSERT INTO secrets (name, type, content, ownerID, keyID)
		VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(insertSQL, s.Name, s.Type, s.Content, s.UserId, s.KeyId)
	return err
}

// DeleteSecret удаляет запись по ID
func DeleteSecret(db *sql.DB, s *proto.DeleteSecretRequest) error {
	deleteSQL := `
		DELETE FROM secrets
		WHERE name = ? AND secrets.ownerID = ?;
		`
	_, err := db.Exec(deleteSQL, s.Name, s.UserId)
	return err
}

func LoginUser(db *sql.DB, userLogin string, userPassword string) (int, error) {
	var userID int
	if userLogin == "" || userPassword == "" {
		err := errors.New("login or password is empty")
		if err != nil {
			logger.Warnf("LoginUser: " + err.Error())
			return -1, err
		}
	}
	selectUserWithPass := `SELECT userID FROM users WHERE userLogin = ? AND userPassword = ?`

	hash := md5.Sum([]byte(userPassword))
	hashedPass := hex.EncodeToString(hash[:])

	result := db.QueryRow(selectUserWithPass, userLogin, hashedPass)
	switch err := result.Scan(&userID); err {
	case pgx.ErrNoRows:
		return -1, nil
	case nil:
		return userID, nil
	case err:
		logger.Warnf("Query LoginUser: " + err.Error())
		return -1, nil
	}
	return userID, nil
}

func GetSecretsList(db *sql.DB, userID int) ([]*proto.SecretsList, error) {
	var secrets []*proto.SecretsList

	if userID <= 0 {
		err := errors.New("wrong userID")
		if err != nil {
			logger.Warnf("LoginUser: " + err.Error())
			return secrets, err
		}
	}

	rows, err := db.Query("SELECT name, type FROM Secrets WHERE ownerID = ?", userID)
	if err != nil {
		return secrets, err
	}
	defer rows.Close()

	for rows.Next() {
		var secret proto.SecretsList
		if err := rows.Scan(&secret.Name, &secret.Type); err != nil {
			log.Fatal(err)
		}
		secrets = append(secrets, &secret)
	}

	if err := rows.Err(); err != nil {
		return secrets, err
	}

	return secrets, nil
}

func GetSecret(db *sql.DB, name string, userID int) (*proto.GetSecretResponse, error) {
	var secret proto.GetSecretResponse

	if userID <= 0 {
		err := errors.New("wrong userID")
		if err != nil {
			logger.Warnf("GetSecret: " + err.Error())
			return &secret, err
		}
	}

	result := db.QueryRow("SELECT type, content, version, keyID FROM Secrets WHERE name = ? AND ownerID = ?", name, userID)
	switch err := result.Scan(&secret.Type, &secret.Content, &secret.Version, &secret.KeyId); err {
	case pgx.ErrNoRows:
		return &secret, nil
	case nil:
		return &secret, nil
	case err:
		logger.Warnf("Query GetSecret: " + err.Error())
		return &secret, err
	}
	return &secret, nil
}

// GetAESKey получает AES key
func GetAESKey(db *sql.DB, ownerID int64, cryptoCert string) (int64, string, error) {
	var KeyID int64
	var keyAES string

	// Используем QueryRow для получения первого значения
	err := db.QueryRow("SELECT KeyID, KeyAES FROM Keys WHERE ownerID = ?", ownerID).Scan(&KeyID, &keyAES)
	if err != nil {
		if err == sql.ErrNoRows {
			//Нет записей с таким ownerID
			// Генерация AES-ключа
			aesKey, err := service.GenerateRandom(32) //generate a random 32 byte key for AES-256
			if err != nil {
				logger.Warnf("Ошибка генерации ключа:" + err.Error())
				return 0, "", err
			}
			// Шифрование AES-ключа с помощью RSA
			keyAES, err = crypt.Encrypt(cryptoCert, hex.EncodeToString(aesKey))
			if err != nil {
				logger.Warnf("Ошибка шифрования AES ключа:" + err.Error())
				return 0, "", err
			}
			err = SaveAESKey(db, ownerID, keyAES)
			if err != nil {
				logger.Warnf("Ошибка записи в таблицу AES ключа:" + err.Error())
				return 0, "", err
			}
		} else {
			logger.Warnf(err.Error())
			return 0, "", err
		}
	}
	return KeyID, keyAES, nil
}

// GetAESKey получает AES key по ID
func GetAESKeyByID(db *sql.DB, ownerID int64, keyID int64) (string, error) {
	var keyAES string

	// Используем QueryRow для получения первого значения
	err := db.QueryRow("SELECT KeyAES FROM Keys WHERE ownerID = ? AND KeyID = ?", ownerID, keyID).Scan(&keyAES)
	if err != nil {
		if err == sql.ErrNoRows {
			//Нет записей
			logger.Warnf(err.Error())
			return "", err
		} else {
			logger.Warnf(err.Error())
			return "", err
		}
	}
	return keyAES, nil
}

// SaveAESKey записывает AES key
func SaveAESKey(db *sql.DB, ownerID int64, keyAES string) error {
	_, err := db.Exec("INSERT INTO Keys (KeyAES, ownerID) VALUES (?, ?)", keyAES, ownerID)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// SyncData cинхронизирует данные между клиентом и сервером
func SyncData(sqliteDB *sql.DB, KeeperClient proto.KeeperClient, NotifyCtx context.Context, ownerID int64) {
	syncDatabasesKeys(sqliteDB, KeeperClient, NotifyCtx, ownerID)
	syncDatabasesUsers(sqliteDB, KeeperClient, NotifyCtx, ownerID)
	// TODO: sync secrets
}

func syncDatabasesKeys(sqliteDB *sql.DB, KeeperClient proto.KeeperClient, NotifyCtx context.Context, ownerID int64) {

	// синхронизируем данные сервер -> клиент

	// Запрос данных из PostgreSQL
	resp, err := KeeperClient.SyncDataKeys(NotifyCtx, &proto.SyncDataKeysRequest{OwnerID: ownerID})
	if err != nil {
		logger.Warnf("Ошибка при вызове SyncData: " + err.Error())
		return
	}

	// Вставка данных в SQLite
	stmt, err := sqliteDB.Prepare(`
        INSERT INTO Keys (KeyID, KeyAES, ownerID, timestamp) 
        VALUES (?, ?, ?, ?) 
        ON CONFLICT (KeyID) 
        DO UPDATE SET KeyAES = EXCLUDED.KeyAES, timestamp = EXCLUDED.timestamp 
        WHERE EXCLUDED.timestamp > Keys.timestamp
    `)
	if err != nil {
		logger.Warnf("Ошибка подготовки запроса в SQLite: " + err.Error())
		return
	}
	defer stmt.Close()

	for _, key := range resp.Keys {
		if _, err := stmt.Exec(key.KeyID, key.KeyAES, key.OwnerID, key.Timestamp); err != nil {
			logger.Warnf("Ошибка вставки/обновленияв SQLite, KeyID: " + err.Error())
		}
	}

	// синхронизируем данные клиент -> сервер

	// Запрос данных из SQLite
	rows, err := sqliteDB.Query("SELECT KeyID, KeyAES, ownerID, timestamp FROM Keys WHERE ownerID = ?", ownerID)
	if err != nil {
		logger.Warnf("Ошибка извлечения данных из SQLite: " + err.Error())
		return
	}
	defer rows.Close()

	var keys []*proto.KeyData
	for rows.Next() {
		var key proto.KeyData
		if err := rows.Scan(&key.KeyID, &key.KeyAES, &key.OwnerID, &key.Timestamp); err != nil {
			logger.Warnf("Ошибка сканирования данных: " + err.Error())
			continue
		}
		keys = append(keys, &key)
	}

	// Отправка данных на сервер (в PostgreSQL)
	_, err = KeeperClient.PushDataKeys(NotifyCtx, &proto.PushDataKeysRequest{Keys: keys})
	if err != nil {
		logger.Warnf("Ошибка при отправке данных на сервер: " + err.Error())
	}

}

func syncDatabasesUsers(sqliteDB *sql.DB, KeeperClient proto.KeeperClient, NotifyCtx context.Context, ownerID int64) {

	// синхронизируем данные сервер -> клиент

	// Запрос данных из PostgreSQL
	resp, err := KeeperClient.SyncDataUsers(NotifyCtx, &proto.SyncDataUserRequest{OwnerID: ownerID})
	if err != nil {
		logger.Warnf("Ошибка при вызове SyncData: " + err.Error())
		return
	}

	// Вставка данных в SQLite
	stmt, err := sqliteDB.Prepare(`
        INSERT INTO Users (userID, userLogin, userPassword) 
        VALUES (?, ?, ?)
    `)
	if err != nil {
		logger.Warnf("Ошибка подготовки запроса в SQLite: " + err.Error())
		return
	}
	defer stmt.Close()

	for _, user := range resp.Users {
		if _, err := stmt.Exec(user.UserID, user.UserLogin, user.UserPassword); err != nil {
			logger.Warnf("Ошибка вставки/обновленияв SQLite, KeyID: " + err.Error())
		}
	}

}
