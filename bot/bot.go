package bot

import (
	"fmt"
	"time"
	"timeTableProject/db"

	"gorm.io/gorm"
)

// Функция для обработки команды бронирования
func HandleBookCommand(database *gorm.DB, userID uint, timeSlotString string) {
	// Парсим время
	timeSlot, err := time.Parse("15:04", timeSlotString) // Предполагаем формат "HH:MM"
	if err != nil {
		fmt.Println("Invalid time format:", err)
		return
	}

	// Пытаемся создать запись
	err = db.BookAppointment(database, userID, timeSlot)
	if err != nil {
		fmt.Println("Failed to book appointment:", err)
		return
	}

	fmt.Println("Appointment booked successfully at", timeSlot)
}
