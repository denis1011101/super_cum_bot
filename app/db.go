package app

import (
	"database/sql"
	"log"
	"os"
    "io"
	"time"
    "path/filepath"
	"fmt"
    "sync"

	_ "github.com/mattn/go-sqlite3"
)

const sqliteTimestampLayout = "2006-01-02 15:04:05Z07:00"

// InitDB инициализирует базу данных
func InitDB() (*sql.DB, error) {
    dbDir := "./data"

    // Проверка, существует ли директория базы данных
    if _, err := os.Stat(dbDir); os.IsNotExist(err) {
        // Директория не существует, создаём её
        err = os.MkdirAll(dbDir, os.ModePerm)
        if err != nil {
            log.Printf("Error creating directory: %v", err)
            return nil, err
        }
    }

	dbPath := "./data/pens.db"

    // Проверка, существует ли директория базы данных
    if _, err := os.Stat(dbDir); os.IsNotExist(err) {
        // Директория не существует, создаём её
        err = os.MkdirAll(dbDir, os.ModePerm)
        if err != nil {
            log.Printf("Error creating directory: %v", err)
            return nil, err
        }
    }

	// Проверка, существует ли файл базы данных
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// Файл не существует, создаём базу данных и таблицу
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Printf("Error opening database: %v", err)
			return nil, err
		}

		createTableQuery := `
		CREATE TABLE IF NOT EXISTS pens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pen_name TEXT,
			tg_pen_id INTEGER UNIQUE,
			tg_chat_id INTEGER,
			pen_length INTEGER,
			pen_last_update_at TIMESTAMP,
			handsome_count INTEGER,
			handsome_last_update_at TIMESTAMP,
			unhandsome_count INTEGER,
			unhandsome_last_update_at TIMESTAMP
		);`
		_, err = db.Exec(createTableQuery)
		if err != nil {
			db.Close() // Закрываем базу данных при ошибке
			log.Printf("Error creating table: %v", err)
			return nil, err
		}

		log.Println("Database and table created successfully")
		return db, nil
	}

	// Файл существует, просто открываем базу данных
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return nil, err
	}

	log.Println("Database opened successfully")
	return db, nil
}

// UpdatePenSize обновляет размер пениса в базе данных
func UpdatePenSize(db *sql.DB, tgChatID int64, newSize int) error {
	_, err := db.Exec(`UPDATE pens SET pen_length = ? WHERE tg_chat_id = ?`, newSize, tgChatID)
	if err != nil {
		log.Printf("Error updating pen size: %v", err)
		return err
	}
	log.Printf("Pen size updated successfully for tg_chat_id: %d, new size: %d", tgChatID, newSize)
	return nil
}

// GetUserIDByUsername получает ID пользователя по его username
func GetUserIDByUsername(db *sql.DB, username string) (int, error) {
	var userID int
	err := db.QueryRow("SELECT tg_pen_id FROM pens WHERE pen_name = ?", username).Scan(&userID)
	if err != nil {
		return 0, err
	}
	log.Printf("User ID retrieved successfully for username: %s, user ID: %d", username, userID)
	return userID, nil
}

// GetPenNames получает все значения pen_name из таблицы pens
func GetPenNames(db *sql.DB, chatID int64) ([]Member, error) {
	rows, err := db.Query("SELECT tg_pen_id, pen_name FROM pens WHERE tg_chat_id = ?", chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var member Member
		err := rows.Scan(&member.ID, &member.Name)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

func GetUserPen(db *sql.DB, userID int64, chatID int64) (Pen, error) {
	var currentSize int
	var lastUpdate sql.NullTime
	err := db.QueryRow("SELECT pen_length, pen_last_update_at FROM pens WHERE tg_pen_id = ? AND tg_chat_id = ?", userID, chatID).Scan(&currentSize, &lastUpdate)
	if err != nil {
		log.Printf("Error querying user pen: %v", err)
		return Pen{}, err
	} else {
		log.Printf("User pen retrieved successfully")
	}
	return Pen{currentSize, lastUpdate.Time}, err
}

func UpdateUserPen(db *sql.DB, userID int64, chatID int64, newSize int) {
	_, err := db.Exec("UPDATE pens SET pen_length = ?, pen_last_update_at = ? WHERE tg_pen_id = ? AND tg_chat_id = ?", newSize, time.Now(), userID, chatID)
	if err != nil {
		log.Printf("Error updating pen size and last update time: %v", err)
	} else {
		log.Printf("Successfully updated pen size and last update time for userID: %d, chatID: %d, newSize: %d", userID, chatID, newSize)
	}
}

func GetGigaLastUpdateTime(db *sql.DB, chatID int64) (time.Time, error) {
	var lastUpdateText sql.NullString
	err := db.QueryRow("SELECT MAX(handsome_last_update_at) FROM pens WHERE tg_chat_id = ?", chatID).Scan(&lastUpdateText)
	if err != nil {
		log.Printf("Error querying last update time: %v", err)
	} else {
		log.Printf("Last update time retrieved successfully")
	}
	if lastUpdateText.Valid {
		lastUpdate, err := time.Parse(sqliteTimestampLayout, lastUpdateText.String)
		if err != nil {
			log.Printf("Error parsing last update time: %v", err)
			return time.Time{}, err
		}
		log.Printf("Last update time parsed successfully")
		return lastUpdate, nil
	}
	log.Printf("Last update time is empty")
	return time.Time{}, nil
}

func GetUnhandsomeLastUpdateTime(db *sql.DB, chatID int64) (time.Time, error) {
	var lastUpdateText sql.NullString
	err := db.QueryRow("SELECT MAX(unhandsome_last_update_at) FROM pens WHERE tg_chat_id = ?", chatID).Scan(&lastUpdateText)
	if err != nil {
		log.Printf("Error querying last update time: %v", err)
	} else {
		log.Printf("Last update time retrieved successfully")
	}
	if lastUpdateText.Valid {
		lastUpdate, err := time.Parse(sqliteTimestampLayout, lastUpdateText.String)
		if err != nil {
			log.Printf("Error parsing last update time: %v", err)
			return time.Time{}, err
		}
		log.Printf("Last update time parsed successfully")
		return lastUpdate, nil
	}
	log.Printf("Last update time is empty")
	return time.Time{}, nil
}

func UpdateGiga(db *sql.DB, newSize int, userID int64, chatID int64) {
    _, err := db.Exec("UPDATE pens SET pen_length = ?, handsome_count = handsome_count + 1, handsome_last_update_at = ? WHERE tg_pen_id = ? AND tg_chat_id = ?", newSize, time.Now(), userID, chatID)
	if err != nil {
		log.Printf("Error updating giga count and last_update_at : %v", err)
	} else {
		log.Printf("Successfully updated giga count and last_update_at for userID: %d, chatID: %d, newSize: %d", userID, chatID, newSize)
	}
}

func UpdateUnhandsome(db *sql.DB, newSize int, userID int64, chatID int64) {
	_, err := db.Exec("UPDATE pens SET pen_length = ?, unhandsome_count = unhandsome_count + 1, unhandsome_last_update_at = ? WHERE tg_pen_id = ? AND tg_chat_id = ?", newSize, time.Now(), userID, chatID)
    if err != nil {
        log.Printf("Error updating unhandsome count and last_update_at: %v", err)
    } else {
        log.Printf("Successfully updated unhandsome count and last_update_at for userID: %d, chatID: %d, newSize: %d", userID, chatID, newSize)
    }
}

// backupDatabase создает резервную копию базы данных SQLite
func backupDatabase() error {
    // Определение пути к файлу резервной копии в корневом каталоге
    source := "./data/pens.db"
    backupDir := "backups"

    // Генерация уникального имени файла резервной копии на основе текущей даты и времени
    timestamp := time.Now().Format("20060102_150405")
    backupFile := filepath.Join(backupDir, "database_backup_"+timestamp+".db")

    // Создание директории для резервной копии, если она не существует
    if _, err := os.Stat(backupDir); os.IsNotExist(err) {
        if err := os.MkdirAll(backupDir, 0755); err != nil {
            log.Fatalf("Cannot create backup directory: %s, error: %v", backupDir, err)
			return err
        }
    }

    // Проверка существования файла базы данных
    if _, err := os.Stat(source); os.IsNotExist(err) {
        log.Fatalf("Database file does not exist: %s", source)
		return fmt.Errorf("database file does not exist: %s", source)
    }

    // Открываем исходный файл базы данных для чтения
    srcFile, err := os.Open(source)
    if err != nil {
        return err
    }
    defer srcFile.Close() // Закрываем файл после завершения функции

    // Создаем файл назначения для записи резервной копии
    dstFile, err := os.Create(backupFile)
    if err != nil {
        return err
    }
    defer dstFile.Close() // Закрываем файл после завершения функции

    // Копируем содержимое исходного файла в файл назначения
    _, err = io.Copy(dstFile, srcFile)
    if err != nil {
        return err
    }

    // Вывод сообщения об успешном создании резервной копии
    log.Printf("Backup created successfully at %s", backupFile)
    return nil // Возвращаем nil, если все операции прошли успешно
}

// StartBackupRoutine запускает процесс резервного копирования
func StartBackupRoutine(db *sql.DB, mutex *sync.Mutex) {
    go func() {
        // Настройка таймера для выполнения раз в день
        ticker := time.NewTicker(24 * time.Hour)
        defer ticker.Stop()

        for range ticker.C {
            // Блокируем базу данных
            mutex.Lock()

            // Выполнение резервного копирования
            if err := backupDatabase(); err != nil {
                log.Printf("Ошибка при резервном копировании базы данных: %v", err)
            } else {
                log.Println("Резервное копирование завершено успешно")
            }

            // Разблокируем базу данных
            mutex.Unlock()

            // Задержка выполнения на 24 часа
            log.Println("Ожидание 24 часов перед следующим резервным копированием...")
        }
    }()
}
