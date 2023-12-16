package linkgen

import (
	"math/rand"
)

// Таблица символов для генерации уникального ID.
var char = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_"

// Функция генерирует ID в размере 10 символов используя цифры, латинские буквы
// и нижнее подчёркивание.
func GenerateShortURL() string {
	shortUrl := ""
	for i := 0; i < 10; i++ {
		shortUrl += string(char[rand.Intn(63)])
	}

	return shortUrl
}
