package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

	// Importamos nuestros paquetes internos
	"printer-service/internal/utils"
	"printer-service/internal/winprint"
)

// Modelos JSON
type PrintRequest struct {
	PrinterType string `json:"type"`       // "local" o "network"
	Identifier  string `json:"identifier"` // Nombre (local) o IP (network)
	Content     struct {
		Title string `json:"title"`
		Items []struct {
			Name  string  `json:"name"`
			Price float64 `json:"price"`
		} `json:"items"`
		Total float64 `json:"total"`
	} `json:"content"`
}

// --- FUNCIÓN HELPER DE ESCANEO DE RED (Ya la tenías, simplificada) ---
func scanNetwork(baseIP string) []map[string]interface{} {
	var results []map[string]interface{}
	var m sync.Mutex
	var wg sync.WaitGroup

	for i := 1; i < 255; i++ {
		wg.Add(1)
		go func(suffix int) {
			defer wg.Done()
			ip := fmt.Sprintf("%s%d", baseIP, suffix)
			conn, err := net.DialTimeout("tcp", ip+":9100", 300*time.Millisecond)
			if err == nil {
				conn.Close()
				m.Lock()
				results = append(results, map[string]interface{}{
					"type": "network", "name": "Net Printer", "ip": ip, "port": 9100,
				})
				m.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return results
}

func main() {
	app := fiber.New()
	app.Use(cors.New())

	// 1. Endpoint Discover (Mezcla Local + Red)
	app.Get("/api/discover", func(c fiber.Ctx) error {
		// A. Locales (Windows USB)
		localNames, _ := winprint.ListLocalPrinters()
		localList := []map[string]interface{}{}
		for _, name := range localNames {
			localList = append(localList, map[string]interface{}{
				"type": "local", "name": name, "detail": "USB/Windows Spooler",
			})
		}

		// B. Red
		netList := scanNetwork("192.168.18.") // Ajusta tu red base aquí

		return c.JSON(fiber.Map{
			"local":   localList,
			"network": netList,
		})
	})

	// 2. Endpoint Print
	app.Post("/api/print", func(c fiber.Ctx) error {
		var req PrintRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(400).SendString("Bad Request")
		}

		// A. Construir el Ticket
		builder := utils.NewTicketBuilder()

		// Logo (Asumiendo que lo cargaste en la memoria de la impresora con la utilidad de 3nStar)
		// builder.PrintNVLogo(1)

		builder.AlignCenter()
		builder.SetBold(true)
		builder.AddTextLn(req.Content.Title)
		builder.SetBold(false)
		builder.AddTextLn("--------------------------------")
		builder.AlignLeft()

		for _, item := range req.Content.Items {
			line := fmt.Sprintf("%-20s %8.2f", item.Name, item.Price)
			builder.AddTextLn(line)
		}

		builder.AddTextLn("--------------------------------")
		builder.AlignRight()
		builder.SetBold(true)
		builder.AddTextLn(fmt.Sprintf("TOTAL: S/ %.2f", req.Content.Total))

		builder.Feed(3)
		builder.Cut()

		data := builder.GetBytes()

		// B. Enviar a Impresora
		var err error
		if req.PrinterType == "network" {
			// Enviar por socket TCP (IP:9100)
			conn, e := net.Dial("tcp", fmt.Sprintf("%s:9100", req.Identifier))
			if e != nil {
				return c.Status(500).SendString("Error conectando a IP: " + e.Error())
			}
			defer conn.Close()
			_, err = conn.Write(data)
		} else {
			// Enviar por USB (Windows Spooler)
			err = winprint.SendBytesToPrinter(req.Identifier, data)
		}

		if err != nil {
			return c.Status(500).SendString("Error imprimiendo: " + err.Error())
		}

		return c.JSON(fiber.Map{"status": "success", "message": "Ticket enviado"})
	})

	app.Listen(":4000")
}
