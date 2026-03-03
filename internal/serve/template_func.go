package serve

import (
	"fmt"
	"html/template"
	"reflect"
	"strconv"
	"time"

	"github.com/ad9311/ninete/internal/handlers"
	"github.com/ad9311/ninete/internal/prog"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"currency":         currency,
		"sumAmount":        sumAmount,
		"sumTotal":         sumTotal,
		"timeStamp":        timeStamp,
		"sortURL":          sortURL,
		"pageURL":          pageURL,
		"pageRange":        pageRange,
		"filterURL":        filterURL,
		"dateRangeOptions": handlers.DateRangeOptions,
		"add":              func(a, b int) int { return a + b },
		"sub":              func(a, b int) int { return a - b },
	}
}

func currency(v uint64) string {
	p := message.NewPrinter(language.AmericanEnglish)

	return p.Sprintf("$%.2f", float64(v)/100.0)
}

func timeStamp(v int64) string {
	return prog.UnixToStringDate(v, time.DateOnly)
}

func sumAmount(rows any) uint64 {
	value := reflect.ValueOf(rows)
	if !value.IsValid() || value.Kind() != reflect.Slice {
		return 0
	}

	var total uint64
	for i := 0; i < value.Len(); i++ {
		item := value.Index(i)
		if item.Kind() == reflect.Pointer {
			if item.IsNil() {
				continue
			}
			item = item.Elem()
		}
		if item.Kind() != reflect.Struct {
			continue
		}

		amount := item.FieldByName("Amount")
		if !amount.IsValid() {
			continue
		}

		total += amount.Uint()
	}

	return total
}

func sumTotal(rows any) uint64 {
	value := reflect.ValueOf(rows)
	if !value.IsValid() || value.Kind() != reflect.Slice {
		return 0
	}

	var total uint64
	for i := 0; i < value.Len(); i++ {
		item := value.Index(i)
		if item.Kind() == reflect.Pointer {
			if item.IsNil() {
				continue
			}
			item = item.Elem()
		}
		if item.Kind() != reflect.Struct {
			continue
		}

		field := item.FieldByName("Total")
		if !field.IsValid() {
			continue
		}

		total += field.Uint()
	}

	return total
}

func filterParams(pg handlers.PaginationData) string {
	params := ""
	if pg.CategoryID > 0 {
		params += fmt.Sprintf("&category_id=%d", pg.CategoryID)
	}

	if pg.DateRange != "" {
		params += "&date_range=" + pg.DateRange
	}

	if pg.Done != "" {
		params += "&done=" + pg.Done
	}

	if pg.Priority > 0 {
		params += fmt.Sprintf("&priority=%d", pg.Priority)
	}

	return params
}

func sortURL(basePath, field string, pg handlers.PaginationData) string {
	order := "ASC"
	if pg.SortField == field && pg.SortOrder == "ASC" {
		order = "DESC"
	}

	return fmt.Sprintf("%s?sort_field=%s&sort_order=%s&per_page=%d&page=1",
		basePath, field, order, pg.PerPage) + filterParams(pg)
}

func pageURL(basePath string, page int, pg handlers.PaginationData) string {
	return fmt.Sprintf("%s?sort_field=%s&sort_order=%s&per_page=%d&page=%d",
		basePath, pg.SortField, pg.SortOrder, pg.PerPage, page) + filterParams(pg)
}

func filterURL(basePath string, pg handlers.PaginationData, key, value string) string {
	categoryID := pg.CategoryID
	dateRange := pg.DateRange
	done := pg.Done
	priority := pg.Priority

	switch key {
	case "category_id":
		categoryID, _ = strconv.Atoi(value)
	case "date_range":
		dateRange = value
	case "done":
		done = value
	case "priority":
		priority, _ = strconv.Atoi(value)
	}

	base := fmt.Sprintf("%s?sort_field=%s&sort_order=%s&per_page=%d&page=1",
		basePath, pg.SortField, pg.SortOrder, pg.PerPage)
	if categoryID > 0 {
		base += fmt.Sprintf("&category_id=%d", categoryID)
	}

	if dateRange != "" {
		base += "&date_range=" + dateRange
	}

	if done != "" {
		base += "&done=" + done
	}

	if priority > 0 {
		base += fmt.Sprintf("&priority=%d", priority)
	}

	return base
}

func pageRange(totalPages, currentPage int) []int {
	if totalPages <= 0 {
		return nil
	}

	start := max(currentPage-2, 1)
	end := min(start+4, totalPages)
	start = max(end-4, 1)

	pages := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}

	return pages
}
