// Command gen-gmp-form-template builds templates/gmp_form.xlsx, the layout
// used by the GMP inspection form export, from a filled reference workbook
// (GMP_Ref.xlsx, sheet "Hal 00"). It keeps the header, one empty prototype
// data row styled like a finding row, the total row, the legend and the
// signature block, and clears every inspection value from the reference.
//
//	go run ./scripts/gen-gmp-form-template -ref /path/GMP_Ref.xlsx
package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	sheet     = "Hal 00"
	dataStart = 10
)

func main() {
	refPath := flag.String("ref", "../frontend/public/docs/GMP_Ref.xlsx", "filled reference workbook")
	output := flag.String("output", "templates/gmp_form.xlsx", "template to write")
	findingRow := flag.Int("finding-row", 24, "a reference row that holds a finding (styles for J, K, M)")
	flag.Parse()

	f, err := excelize.OpenFile(*refPath)
	if err != nil {
		log.Fatalf("open reference: %v", err)
	}
	defer f.Close()

	totalRow, err := findRow(f, "Total Score IPGMP")
	if err != nil {
		log.Fatal(err)
	}

	// Take the finding styles before the rows holding them are removed.
	findingStyles := map[string]int{}
	// G is taken from a non-finding row: G24 carries the red "0" fill.
	for _, col := range []string{"I", "J", "K", "L", "M"} {
		style, err := f.GetCellStyle(sheet, fmt.Sprintf("%s%d", col, *findingRow))
		if err != nil {
			log.Fatal(err)
		}
		findingStyles[col] = style
	}

	// Drop every data row except the prototype (row 10).
	for row := totalRow - 1; row > dataStart; row-- {
		if err := f.RemoveRow(sheet, row); err != nil {
			log.Fatalf("remove row %d: %v", row, err)
		}
	}

	// Prototype row: only E:F stays merged; vertical group merges are
	// rebuilt by the exporter for each inspection.
	merges, err := f.GetMergeCells(sheet)
	if err != nil {
		log.Fatal(err)
	}
	for _, m := range merges {
		start, end := m.GetStartAxis(), m.GetEndAxis()
		_, r1, _ := excelize.CellNameToCoordinates(start)
		_, r2, _ := excelize.CellNameToCoordinates(end)
		if r1 <= dataStart && r2 >= dataStart && !(start == fmt.Sprintf("E%d", dataStart) && end == fmt.Sprintf("F%d", dataStart)) {
			if err := f.UnmergeCell(sheet, start, end); err != nil {
				log.Fatal(err)
			}
		}
	}
	_ = f.MergeCell(sheet, fmt.Sprintf("E%d", dataStart), fmt.Sprintf("F%d", dataStart))

	for _, col := range strings.Split("A B C D E F G H I J K L M", " ") {
		cell := fmt.Sprintf("%s%d", col, dataStart)
		if err := f.SetCellValue(sheet, cell, nil); err != nil {
			log.Fatal(err)
		}
		if style, ok := findingStyles[col]; ok {
			_ = f.SetCellStyle(sheet, cell, cell, style)
		}
	}
	// Header values are per inspection.
	for _, cell := range []string{"C5", "C6", "C7"} {
		_ = f.SetCellValue(sheet, cell, nil)
	}
	_ = f.SetCellValue(sheet, "C4", "INSPEKSI GMP (BASIC)")

	// Total row moved up to dataStart+1; the exporter writes its formulas.
	total := dataStart + 1
	for _, col := range []string{"G", "I", "M"} {
		_ = f.SetCellFormula(sheet, fmt.Sprintf("%s%d", col, total), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, total), nil)
	}

	if err := f.SaveAs(*output); err != nil {
		log.Fatalf("save: %v", err)
	}
	// excelize never drops unused shared strings, so the reference's
	// inspection text would ship inside every export: rewrite the table
	// with only the strings the remaining cells use.
	if err := compactSharedStrings(*output); err != nil {
		log.Fatalf("compact shared strings: %v", err)
	}
	fmt.Printf("wrote %s (prototype row %d, total row %d)\n", *output, dataStart, total)
}

func findRow(f *excelize.File, prefix string) (int, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return 0, err
	}
	for i, row := range rows {
		if len(row) > 0 && strings.HasPrefix(row[0], prefix) {
			return i + 1, nil
		}
	}
	return 0, fmt.Errorf("row starting with %q not found", prefix)
}

var (
	sharedRefRe = regexp.MustCompile(`(<c [^>]*t="s"[^>]*>(?:<f[^<]*</f>)?<v>)(\d+)(</v>)`)
	siRe        = regexp.MustCompile(`(?s)<si>.*?</si>`)
	sstHeadRe   = regexp.MustCompile(`(?s)^(.*?<sst[^>]*?)(?: count="\d+")?(?: uniqueCount="\d+")?(>)`)
)

func compactSharedStrings(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	files := map[string][]byte{}
	var order []*zip.File
	for _, zf := range zr.File {
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return err
		}
		files[zf.Name] = b
		order = append(order, zf)
	}
	sst, ok := files["xl/sharedStrings.xml"]
	if !ok {
		return nil
	}
	items := siRe.FindAll(sst, -1)

	remap := map[int]int{}
	var kept [][]byte
	for name, b := range files {
		if !strings.HasPrefix(name, "xl/worksheets/sheet") {
			continue
		}
		files[name] = sharedRefRe.ReplaceAllFunc(b, func(m []byte) []byte {
			parts := sharedRefRe.FindSubmatch(m)
			old, _ := strconv.Atoi(string(parts[2]))
			idx, seen := remap[old]
			if !seen {
				idx = len(kept)
				remap[old] = idx
				kept = append(kept, items[old])
			}
			return []byte(string(parts[1]) + strconv.Itoa(idx) + string(parts[3]))
		})
	}

	head := sstHeadRe.FindSubmatch(sst)
	var out bytes.Buffer
	out.Write(head[1])
	fmt.Fprintf(&out, ` count="%d" uniqueCount="%d">`, len(kept), len(kept))
	for _, item := range kept {
		out.Write(item)
	}
	out.WriteString("</sst>")
	files["xl/sharedStrings.xml"] = out.Bytes()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, zf := range order {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: zf.Name, Method: zip.Deflate, Modified: zf.Modified})
		if err != nil {
			return err
		}
		if _, err := w.Write(files[zf.Name]); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
