package parser

import (
	"bytes"
	"context"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestCSVParserSuccess(t *testing.T) {
	t.Parallel()

	parser := csvParser{}
	rows, err := parser.Parse(context.Background(), []byte("name,category,price,description,store_link\nBook,Books,10.00,Gift,https://example.com/book\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Parse() rows = %d, want 1", len(rows))
	}
	if rows[0].Name != "Book" {
		t.Fatalf("Parse() name = %q, want %q", rows[0].Name, "Book")
	}
}

func TestCSVParserBrokenFile(t *testing.T) {
	t.Parallel()

	parser := csvParser{}
	if _, err := parser.Parse(context.Background(), []byte("\"name\",\"category\"\n\"broken")); err == nil {
		t.Fatal("Parse() expected error for broken csv")
	}
}

func TestJSONParserSuccess(t *testing.T) {
	t.Parallel()

	parser := jsonParser{}
	rows, err := parser.Parse(context.Background(), []byte(`[{"name":"Book","category":"Books","price":"10.00","description":"Gift","store_link":"https://example.com/book"}]`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Parse() rows = %d, want 1", len(rows))
	}
}

func TestJSONParserBrokenFile(t *testing.T) {
	t.Parallel()

	parser := jsonParser{}
	if _, err := parser.Parse(context.Background(), []byte(`{"items":`)); err == nil {
		t.Fatal("Parse() expected error for broken json")
	}
}

func TestXLSXParserSuccess(t *testing.T) {
	t.Parallel()

	workbook := excelize.NewFile()
	sheet := workbook.GetSheetName(0)
	if err := workbook.SetSheetRow(sheet, "A1", &[]string{"name", "category", "price", "description", "store_link"}); err != nil {
		t.Fatalf("SetSheetRow() error = %v", err)
	}
	if err := workbook.SetSheetRow(sheet, "A2", &[]string{"Book", "Books", "10.00", "Gift", "https://example.com/book"}); err != nil {
		t.Fatalf("SetSheetRow() error = %v", err)
	}

	var buffer bytes.Buffer
	if err := workbook.Write(&buffer); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	parser := xlsxParser{}
	rows, err := parser.Parse(context.Background(), buffer.Bytes())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Parse() rows = %d, want 1", len(rows))
	}
}

func TestXLSXParserBrokenFile(t *testing.T) {
	t.Parallel()

	parser := xlsxParser{}
	if _, err := parser.Parse(context.Background(), []byte("not-an-xlsx")); err == nil {
		t.Fatal("Parse() expected error for broken xlsx")
	}
}

func TestCSVParserMissingRequiredHeader(t *testing.T) {
	t.Parallel()

	parser := csvParser{}
	if _, err := parser.Parse(context.Background(), []byte("name,price\nBook,10.00\n")); err == nil {
		t.Fatal("Parse() expected error for missing header")
	}
}
