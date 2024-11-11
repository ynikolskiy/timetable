package auth

import (
	"timeTableProject/db"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthenticateUser проверяет логин и пароль пользователя
func AuthenticateUser(database *gorm.DB, username, password string) (bool, error) {
	var user db.User

	// Ищем пользователя по логину
	result := database.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return false, nil // Пользователь не найден
		}
		return false, result.Error // Возвращаем ошибку, если произошла другая ошибка
	}

	// Сравниваем введённый пароль с хешем, сохранённым в базе
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return false, nil // Если пароли не совпадают, возвращаем false
	}

	return true, nil // Успешная аутентификация
}
