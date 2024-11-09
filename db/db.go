package db

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"` // Здесь можно сохранить хеш пароля
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err // Возвращаем ошибку, если хеширование не удалось
	}

	// Создаем пользователя с захешированным паролем
	user := User{
		Username: username,
		Password: string(hashedPassword), // Сохраняем хеш, а не сам пароль
	}

	// Выполняем вставку записи в базу данных
	result := db.Create(&user)

	return result.Error // Возвращаем ошибку, если она возникнет
}

func BookAppointment(db *gorm.DB, userID uint, timeSlot time.Time) error {
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
