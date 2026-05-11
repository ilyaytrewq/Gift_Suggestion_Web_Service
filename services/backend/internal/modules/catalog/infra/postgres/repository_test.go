package postgres

import (
	"strings"
	"testing"

	catalogusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/usecase"
)

func TestBuildGiftWhereSearchIncludesCategoryName(t *testing.T) {
	t.Parallel()

	where, args := buildGiftWhere(catalogusecase.GiftFilter{Search: "Книги"})

	if !strings.Contains(where, "c.name ILIKE") {
		t.Fatalf("where = %q, expected category name search", where)
	}
	if len(args) != 1 {
		t.Fatalf("args len = %d, want 1", len(args))
	}
	if args[0] != "%Книги%" {
		t.Fatalf("args[0] = %q, want %q", args[0], "%Книги%")
	}
}

func TestEscapeLikePattern(t *testing.T) {
	t.Parallel()

	if got := escapeLikePattern(`100% off_sale`); got != `100\% off\_sale` {
		t.Fatalf("escapeLikePattern() = %q", got)
	}
}
