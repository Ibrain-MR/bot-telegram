package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// Estructuras para manejar el webhook de Telegram
type TelegramUpdate struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Chat struct {
	ID int64 `json:"id"`
}

var (
	bot *tgbotapi.BotAPI
)

// main() es el punto de entrada para un servidor tradicional
func main() {
	// Carga las variables de entorno desde el archivo .env
	// Esto es útil para desarrollo local. En Railway, las variables se inyectan directamente.
	err := godotenv.Load()
	if err != nil {
		log.Println("Advertencia: No se pudo cargar el archivo .env. Asegúrate de que las variables de entorno están configuradas manualmente.")
	}

	// Obtén el token del bot de las variables de entorno
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("Error: La variable de entorno TELEGRAM_BOT_TOKEN no está configurada.")
	}

	// Inicializa el bot de Telegram
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Bot autorizado en la cuenta %s", bot.Self.UserName)

	// Inicializa el router de Gin
	router := gin.Default()

	// Ruta para el "Hello World" en la raíz
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World from your Go Bot!",
		})
	})

	// Define la ruta del webhook.
	router.POST("/webhook", handleTelegramWebhook)

	// Obtén el puerto de la variable de entorno PORT (usado por Railway)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Puerto por defecto si no se especifica
	}

	log.Printf("Iniciando servidor en el puerto :%s", port)
	// Inicia el servidor de Gin, escuchando en el puerto especificado por la variable de entorno
	log.Fatal(router.Run(":" + port))
}

// handleTelegramWebhook es el handler para los mensajes de Telegram
func handleTelegramWebhook(c *gin.Context) {
	var update TelegramUpdate

	// Se enlaza el JSON del cuerpo de la petición con la estructura
	if err := c.BindJSON(&update); err != nil {
		log.Printf("Error al decodificar JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON no válido"})
		return
	}

	log.Printf("Nuevo mensaje del chat %d: %s", update.Message.Chat.ID, update.Message.Text)

	// Verifica si el mensaje contiene un enlace
	url := update.Message.Text
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		// Envía el mismo enlace de vuelta al usuario
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "¡Enlace recibido! Aquí está el enlace que enviaste: "+url)
		bot.Send(msg)
	} else {
		// Maneja mensajes que no son enlaces, si es necesario
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Por favor, envía un enlace que comience con http:// o https://")
		bot.Send(msg)
	}

	// Responde al webhook de Telegram.
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}