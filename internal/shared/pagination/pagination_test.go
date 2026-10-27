package pagination

import (
	"net/url"
	"testing"
)

func TestParseDefaults(t *testing.T) {
	request, err := Parse(url.Values{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if request.Page != 1 || request.PageSize != 20 || request.Offset() != 0 {
		t.Fatalf("unexpected request: %+v", request)
	}
}

func TestParseRejectsOversizedPage(t *testing.T) {
	_, err := Parse(url.Values{"page_size": []string{"101"}})
	if err == nil {
		t.Fatal("expected oversized page_size to fail")
	}
}
