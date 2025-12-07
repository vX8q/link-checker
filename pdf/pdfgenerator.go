package pdf

import (
	"bytes"
	"fmt"
	"log"

	"github.com/signintech/gopdf"
)

// Generate формирует PDF-байты из переданных данных (map: id -> (link->status)).
// Использует библиотеку gopdf и шрифт `arial.ttf` (шрифт должен быть
// доступен в рабочей директории приложения).
func Generate(data map[int]map[string]string) []byte {
	p := gopdf.GoPdf{}
	p.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	p.AddPage()

	if err := p.AddTTFFont("Arial", "arial.ttf"); err != nil {
		log.Printf("failed to add font: %v", err)
		return nil
	}
	if err := p.SetFont("Arial", "", 12); err != nil {
		log.Printf("failed to set font: %v", err)
		return nil
	}

	y := 10.0

	for id, links := range data {
		for link, status := range links {
			p.SetX(10)
			p.SetY(y)
			p.Cell(nil, fmt.Sprintf("[%d] %s — %s", id, link, status))
			y += 8
		}
	}

	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		log.Printf("failed to write PDF: %v", err)
		return nil
	}

	return buf.Bytes()
}
