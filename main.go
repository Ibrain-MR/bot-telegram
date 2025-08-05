package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// Constantes y variables globales
const (
	screenshotFileName = "screenshot.png"
)

var (
	bot *tgbotapi.BotAPI
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

// Función principal
func main() {
	// Carga las variables de entorno desde el archivo .env
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
	// No uses 'var err' aquí, ya que 'err' ya fue declarado arriba con 'godotenv.Load()'
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Bot autorizado en la cuenta %s", bot.Self.UserName)

	router := gin.Default()
	
	// Define la ruta del webhook. Usa solo la ruta sin el token para mayor claridad.
	router.POST("/webhook", handleTelegramWebhook)

	// Inicia el servidor
	router.Run(":8080")
}

// Handler para el webhook
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
		// Envía una confirmación al usuario de que se está procesando
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "¡Enlace detectado! Tomando una captura de pantalla...")
		bot.Send(msg)
		
		// Inicia la captura de pantalla en una goroutine para no bloquear el webhook
		go takeAndSendScreenshot(url, update.Message.Chat.ID)
	}

	// Responde al webhook de Telegram. Esto es vital para que Telegram sepa que el mensaje fue recibido.
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// takeAndSendScreenshot toma la captura y la envía
func takeAndSendScreenshot(urlstr string, chatID int64) {
	// Elimina cualquier archivo de captura anterior para evitar conflictos
	os.Remove(screenshotFileName)

	// Contexto para chromedp
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	if err := chromedp.Run(ctx,
		// Define el tamaño de la ventana del navegador
		chromedp.EmulateViewport(1920, 1080),
		// Navega a la URL
		chromedp.Navigate(urlstr),
		// Espera 2 segundos para que la página cargue completamente
		chromedp.Sleep(2*time.Second),
		// Toma una captura de pantalla completa
		chromedp.FullScreenshot(&buf, 90),
	); err != nil {
		log.Printf("Error al tomar la captura de pantalla: %v", err)
		errMsg := tgbotapi.NewMessage(chatID, "Ocurrió un error al tomar la captura de pantalla.")
		bot.Send(errMsg)
		return
	}

	// Guarda el buffer de la imagen en un archivo
	if err := os.WriteFile(screenshotFileName, buf, 0644); err != nil {
		log.Printf("Error al guardar la captura de pantalla: %v", err)
		errMsg := tgbotapi.NewMessage(chatID, "Ocurrió un error al guardar la captura de pantalla.")
		bot.Send(errMsg)
		return
	}

	// Crea un objeto para enviar la foto
	photoMsg := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(screenshotFileName))
	photoMsg.Caption = "Captura de pantalla de la URL."

	// Envía la foto al chat de Telegram
	if _, err := bot.Send(photoMsg); err != nil {
		log.Printf("Error al enviar la foto a Telegram: %v", err)
	}
	
	// Limpia el archivo de captura una vez enviado
	os.Remove(screenshotFileName)
}
