// Package poster builds the A4 offer poster as one self-contained HTML page.
// The same HTML is rendered to a PNG by headless Chrome and shown for printing.
package poster

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"
)

const (
	HotDeal      = "hot_deal"
	WeekendOffer = "weekend_offer"
	LangEN       = "en"
	LangAR       = "ar"

	// PNG size of an A4 page at 150 dpi.
	Width  = 1240
	Height = 1754
)

var ErrBadChoice = errors.New("template must be hot_deal or weekend_offer and lang must be en or ar")

func ValidChoice(tpl, lang string) bool {
	return (tpl == HotDeal || tpl == WeekendOffer) && (lang == LangEN || lang == LangAR)
}

// Offer is what the poster needs, already loaded by the caller.
type Offer struct {
	ItemName, ItemNameAr string
	Brand, Unit          string
	RegularPrice         float64
	OfferPrice           float64
	DiscountPct          int
	ValidFrom, ValidTo   string // YYYY-MM-DD
	Status               string // upcoming | active | expired
	DaysLeft             int
	PhotoBytes           []byte // nil when the offer has no photo
	PhotoType            string // image/jpeg | image/png | image/webp
}

type Shop struct {
	Name     string
	Branches []string
	Areas    []string
}

type view struct {
	Template string
	RTL      bool
	Print    bool

	Shop, Initial, Areas string
	Branches             string
	Brand, Item, Unit    string
	PhotoURI             template.URL
	ItemInitial          string

	PriceInt, PriceDec, Was string
	Discount                int
	Valid, Banner           string
	L                       labels
}

type labels struct {
	HotDeal, Weekend, Off, Currency, Available, WasLabel string
}

var en = labels{"Hot deal", "Weekend offer", "OFF", "AED", "Available at", "Was"}
var ar = labels{"عرض ساخن", "عرض نهاية الأسبوع", "خصم", "د.إ", "متوفر في", "السعر السابق"}

var monthsEN = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
var monthsAR = []string{"يناير", "فبراير", "مارس", "أبريل", "مايو", "يونيو", "يوليو", "أغسطس", "سبتمبر", "أكتوبر", "نوفمبر", "ديسمبر"}

func day(s string, months []string) (string, int, int) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return s, 0, 0
	}
	return fmt.Sprintf("%d %s", t.Day(), months[t.Month()-1]), t.Year(), t.Day()
}

func validity(from, to, lang string) string {
	months := monthsEN
	if lang == LangAR {
		months = monthsAR
	}
	f, _, _ := day(from, months)
	t, y, _ := day(to, months)
	if lang == LangAR {
		return fmt.Sprintf("صالح من %s إلى %s %d", f, t, y)
	}
	return fmt.Sprintf("Valid %s – %s %d", f, t, y)
}

func banner(o Offer, lang string) string {
	if o.Status != "active" {
		return ""
	}
	if lang == LangAR {
		switch {
		case o.DaysLeft <= 0:
			return "ينتهي اليوم"
		case o.DaysLeft == 1:
			return "باقي يوم واحد"
		default:
			return fmt.Sprintf("باقي %d أيام", o.DaysLeft)
		}
	}
	switch {
	case o.DaysLeft <= 0:
		return "Ends today"
	case o.DaysLeft == 1:
		return "1 day left"
	default:
		return fmt.Sprintf("%d days left", o.DaysLeft)
	}
}

func initial(s string) string {
	for _, r := range strings.TrimSpace(s) {
		return strings.ToUpper(string(r))
	}
	return "?"
}

func splitPrice(p float64) (string, string) {
	cents := int64(math.Round(p * 100))
	return fmt.Sprintf("%d", cents/100), fmt.Sprintf(".%02d", cents%100)
}

// HTML returns the poster page. print=true sizes the page for A4 paper and
// opens the print dialog; false sizes it as the 1240 x 1754 render canvas.
func HTML(tpl, lang string, o Offer, s Shop, print bool) (string, error) {
	if !ValidChoice(tpl, lang) {
		return "", ErrBadChoice
	}
	rtl := lang == LangAR
	l := en
	if rtl {
		l = ar
	}
	item := o.ItemName
	if rtl && strings.TrimSpace(o.ItemNameAr) != "" {
		item = o.ItemNameAr
	}
	pi, pd := splitPrice(o.OfferPrice)
	v := view{
		Template: tpl, RTL: rtl, Print: print,
		Shop: s.Name, Initial: initial(s.Name),
		Areas:    strings.Join(s.Areas, " · "),
		Branches: strings.Join(s.Branches, " · "),
		Brand:    o.Brand, Item: item, Unit: o.Unit, ItemInitial: initial(item),
		PriceInt: pi, PriceDec: pd, Was: fmt.Sprintf("%.2f", o.RegularPrice),
		Discount: o.DiscountPct, Valid: validity(o.ValidFrom, o.ValidTo, lang),
		Banner: banner(o, lang), L: l,
	}
	if len(o.PhotoBytes) > 0 {
		ct := o.PhotoType
		if ct != "image/jpeg" && ct != "image/png" && ct != "image/webp" {
			ct = "image/jpeg"
		}
		v.PhotoURI = template.URL("data:" + ct + ";base64," + base64.StdEncoding.EncodeToString(o.PhotoBytes))
	}
	var buf bytes.Buffer
	if err := page.Execute(&buf, v); err != nil {
		return "", err
	}
	return buf.String(), nil
}
