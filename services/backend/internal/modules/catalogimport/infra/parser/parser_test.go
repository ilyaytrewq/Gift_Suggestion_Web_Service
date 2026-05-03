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

func TestCSVParserTitleAndStoreURLAliases(t *testing.T) {
	t.Parallel()

	// title → name, store_url → store_link, currency is optional
	input := "gift_id,title,description,category,price,currency,store_url,image_url,age_min,age_max\n" +
		"WB1,Плед,Описание,Books,990,RUB,https://wildberries.ru/1,https://img.example.com/1.jpg,3,99\n"

	p := csvParser{}
	rows, err := p.Parse(context.Background(), []byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("Parse() rows = %d, want 1", len(rows))
	}
	if rows[0].Name != "Плед" {
		t.Fatalf("Name = %q, want Плед", rows[0].Name)
	}
	if rows[0].StoreLink != "https://wildberries.ru/1" {
		t.Fatalf("StoreLink = %q, want wildberries url", rows[0].StoreLink)
	}
	if rows[0].Currency != "RUB" {
		t.Fatalf("Currency = %q, want RUB", rows[0].Currency)
	}
}

func TestJSONParserOffersArray(t *testing.T) {
	t.Parallel()

	input := `[{
		"name":"Кружка","category":"Books","price":"500","description":"Gift",
		"store_link":"https://example.com/1",
		"currency":"RUB",
		"offers":[
			{"store_name":"AliExpress","store_url":"https://aliexpress.com/1","price":"350","currency":"CNY"}
		]
	}]`

	p := jsonParser{}
	rows, err := p.Parse(context.Background(), []byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("Parse() rows = %d, want 1", len(rows))
	}
	if rows[0].Currency != "RUB" {
		t.Fatalf("Currency = %q, want RUB", rows[0].Currency)
	}
	if len(rows[0].ExtraOffers) != 1 {
		t.Fatalf("ExtraOffers = %d, want 1", len(rows[0].ExtraOffers))
	}
	if rows[0].ExtraOffers[0].StoreURL != "https://aliexpress.com/1" {
		t.Fatalf("ExtraOffers[0].StoreURL = %q", rows[0].ExtraOffers[0].StoreURL)
	}
	if rows[0].ExtraOffers[0].Currency != "CNY" {
		t.Fatalf("ExtraOffers[0].Currency = %q, want CNY", rows[0].ExtraOffers[0].Currency)
	}
}
