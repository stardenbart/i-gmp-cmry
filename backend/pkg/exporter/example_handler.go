package exporter

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// ExampleFiberHandler demonstrates how to use the Dynamic Excel Exporter in a Fiber HTTP Handler.
func ExampleFiberHandler(c *fiber.Ctx) error {
	// 1. Load the dynamic JSON configuration
	// In production, you might load this once on app startup and cache it, 
	// or load it dynamically based on the requested template ID.
	config, err := LoadConfig("./templates/template_config.json")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to load template config: %v", err),
		})
	}

	// 2. Prepare the Payload
	// In a real scenario, you fetch this data from your Database/UseCase.
	payload := &ExportPayload{
		TemplateID: config.TemplateID,
		Headers: map[string]interface{}{
			"tanggal": "2026-07-20",
			"area":    "Production Line A",
			"pic":     "John Doe",
		},
		TableData: []TableRow{
			{
				Data: map[string]interface{}{
					"uraian_id":  "UR-001",
					"nilai":      85,
					"keterangan": "Memenuhi standar, namun perlu dibersihkan ulang.",
				},
				// Optional: Path to a local image file downloaded/saved temporarily
				// ImagePath: "./temp/temuan_1.png", 
			},
			{
				Data: map[string]interface{}{
					"uraian_id":  "UR-002",
					"nilai":      60,
					"keterangan": "Tidak memenuhi standar kebersihan GMP.",
				},
			},
		},
	}

	// 3. Generate the Excel file to a buffer
	buf, err := GenerateExcel(config, payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to generate excel: %v", err),
		})
	}

	// 4. Send the buffer as a downloadable file stream
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=\"report_gmp.xlsx\"")

	// Fiber makes it easy to send a byte buffer directly
	return c.SendStream(buf)
}
