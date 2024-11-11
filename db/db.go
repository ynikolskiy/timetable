package db

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var bookingMutex sync.Mutex

type User struct {
	ID           uint           `gorm:"primaryKey"`      // Первичный ключ
	Username     string         `gorm:"unique;not null"` // Имя пользователя, уникальное и не пустое
	PasswordHash string         `gorm:"not null"`        // Хеш пароля, не пустое поле
	IsAdmin      bool           `gorm:"default:false"`   // Указывает, является ли пользователь администратором, по умолчанию false
	CreatedAt    time.Time      // Время создания
	UpdatedAt    time.Time      // Время последнего обновления
	DeletedAt    gorm.DeletedAt `gorm:"index"` // Поле для мягкого удаления
}

type Appointment struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      // Внешний ключ для идентификации пользователя
	TimeSlot  time.Time // Время записи
	CreatedAt time.Time
	UpdatedAt time.Time
}

func InitDB() (*gorm.DB, error) {
	dsn := "user=youruser password=yourpassword dbname=yourdb host=localhost port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&User{}, &Appointment{})

	return db, nil
}

func GetAllUsers(db *gorm.DB) ([]User, error) {
	var users []User
	result := db.Find(&users)
	return users, result.Error
}

func CreateUser(db *gorm.DB, username, password string) error {
	// Хешируем пароль
	PasswordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err // Возвращаем ошибку, если хеширование не удалось
	}

	// Создаем пользователя с захешированным паролем
	user := User{
		Username:     username,
		PasswordHash: string(PasswordHash), // Сохраняем хеш, а не сам пароль
	}

	// Выполняем вставку записи в базу данных
	result := db.Create(&user)

	return result.Error // Возвращаем ошибку, если она возникнет
}

// BookAppointment добавляет новую запись для пользователя на заданный временной интервал
func BookAppointment(db *gorm.DB, userID uint, timeSlot time.Time) error {
	bookingMutex.Lock()         // Блокируем мьютекс перед проверкой и записью
	defer bookingMutex.Unlock() // Разблокируем мьютекс после завершения функции

	// Проверка, что время не забронировано другим пользователем
	var existingAppointment Appointment
	if err := db.Where("time_slot = ?", timeSlot).First(&existingAppointment).Error; err == nil {
		return fmt.Errorf("time slot already booked")
	}

	// Создание новой записи
	appointment := Appointment{
		UserID:   userID,
		TimeSlot: timeSlot,
	}
	result := db.Create(&appointment)
	return result.Error
}

func RescheduleAppointment(db *gorm.DB, appointmentID uint, newTimeSlot time.Time) error {
	bookingMutex.Lock()         // Блокируем мьютекс перед началом операции
	defer bookingMutex.Unlock() // Разблокируем мьютекс в конце

	// Проверка, что новый временной интервал не занят другим пользователем
	var existingAppointment Appointment
	if err := db.Where("time_slot = ?", newTimeSlot).First(&existingAppointment).Error; err == nil {
		return fmt.Errorf("new time slot already booked")
	}

	// Поиск текущей записи по её ID
	var appointment Appointment
	if err := db.First(&appointment, appointmentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("appointment not found")
		}
		return err
	}

	// Обновление временного интервала записи
	appointment.TimeSlot = newTimeSlot
	if err := db.Save(&appointment).Error; err != nil {
		return fmt.Errorf("failed to reschedule appointment: %v", err)
	}

	return nil
}

// CancelAppointment удаляет запись, если это инициирует создатель записи или администратор
func CancelAppointment(db *gorm.DB, appointmentID uint, userID uint, isAdmin bool) error {
	// Поиск записи по её ID
	var appointment Appointment
	if err := db.First(&appointment, appointmentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("appointment not found")
		}
		return err
	}

	// Проверка прав на удаление
	if appointment.UserID != userID && !isAdmin {
		return fmt.Errorf("permission denied: only the creator or an admin can cancel the appointment")
	}

	// Удаление записи
	if err := db.Delete(&appointment).Error; err != nil {
		return fmt.Errorf("failed to cancel appointment: %v", err)
	}

	return nil
}
