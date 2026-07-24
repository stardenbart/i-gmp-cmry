package exporter

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// GenerateExcel dynamically generates an Excel file from a template and payload.
func GenerateExcel(config *TemplateConfig, payload *ExportPayload) (*bytes.Buffer, error) {
	// 1. Open the template file
	f, err := excelize.OpenFile(config.TemplateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open template file %s: %w", config.TemplateFile, err)
	}
	defer f.Close()

	sheet := config.SheetName

	// 2. Map Headers
	for key, cellPos := range config.HeaderMapping {
		if val, ok := payload.Headers[key]; ok {
			if err := f.SetCellValue(sheet, cellPos, val); err != nil {
				return nil, fmt.Errorf("failed to set header %s at %s: %w", key, cellPos, err)
			}
		}
	}

	// 3. Map Table Data
	currentRow := config.TableMapping.StartRow
	
	// Pre-duplicate template rows to preserve styling and push footers down
	for i := 1; i < len(payload.TableData); i++ {
		if err := f.DuplicateRow(sheet, config.TableMapping.StartRow + i - 1); err != nil {
			return nil, fmt.Errorf("failed to duplicate template row: %w", err)
		}
	}

	for _, row := range payload.TableData {
		// Insert data into mapped columns
		for key, colLetter := range config.TableMapping.ColumnMap {
			if val, ok := row.Data[key]; ok {
				cell := colLetter + strconv.Itoa(currentRow)
				if err := f.SetCellValue(sheet, cell, val); err != nil {
					return nil, fmt.Errorf("failed to set cell %s: %w", cell, err)
				}
			}
		}

		// Insert image if path is provided and image column is mapped
		if row.ImagePath != "" && config.TableMapping.ImageColumn != "" {
			cell := config.TableMapping.ImageColumn + strconv.Itoa(currentRow)
			
			// Optional: You can configure image size/format here if needed
			// Note: Adjust the image size according to the cell height/width in your template
			picFormat := &excelize.GraphicOptions{
				AutoFit: true,
				OffsetX: 5,
				OffsetY: 5,
			}

			if err := f.AddPicture(sheet, cell, row.ImagePath, picFormat); err != nil {
				// Don't fail the entire export if one image fails to load, just continue
				// but in a production environment, logging the error might be useful.
				fmt.Printf("Warning: failed to add picture %s at %s: %v\n", row.ImagePath, cell, err)
			}
		}

		currentRow++
	}

	// 4. Write to buffer
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel to buffer: %w", err)
	}

	return buf, nil
}
