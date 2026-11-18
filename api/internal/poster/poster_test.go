package poster

import (
	"strings"
	"testing"
)

var testOffer = Offer{
	ItemName: "Fresh Cream", ItemNameAr: "كريمة طازجة", Brand: "Amul", Unit: "200 ml",
	RegularPrice: 10, OfferPrice: 8, DiscountPct: 20,
	ValidFrom: "2026-11-18", ValidTo: "2026-11-28", Status: "active", DaysLeft: 10,
}

var testShop = Shop{Name: "Al Noor Fresh Mart", Branches: []string{"Deira City Centre"}, Areas: []string{"Deira"}}

func TestValidChoice(t *testing.T) {
	if !ValidChoice(HotDeal, LangEN) || !ValidChoice(WeekendOffer, LangAR) {
		t.Error("known template and language pairs must be valid")
	}
	if ValidChoice("flyer", LangEN) || ValidChoice(HotDeal, "fr") || ValidChoice("", "") {
		t.Error("unknown choices must be rejected")
	}
}

func TestHTMLForEveryTemplateAndLanguage(t *testing.T) {
	for _, tpl := range []string{HotDeal, WeekendOffer} {
		for _, lang := range []string{LangEN, LangAR} {
			out, err := HTML(tpl, lang, testOffer, testShop, false)
			if err != nil {
				t.Fatalf("%s/%s: %v", tpl, lang, err)
			}
			for _, want := range []string{"Al Noor Fresh Mart", "20", "8", ".00"} {
				if !strings.Contains(out, want) {
					t.Errorf("%s/%s: poster is missing %q", tpl, lang, want)
				}
			}
			if lang == LangAR && !strings.Contains(out, "كريمة طازجة") {
				t.Errorf("%s/ar: poster should use the Arabic item name", tpl)
			}
			if lang == LangEN && !strings.Contains(out, "Fresh Cream") {
				t.Errorf("%s/en: poster should use the English item name", tpl)
			}
		}
	}
}

func TestHTMLRejectsBadChoice(t *testing.T) {
	if _, err := HTML("nope", LangEN, testOffer, testShop, false); err == nil {
		t.Fatal("expected an error for an unknown template")
	}
}

func TestHTMLEscapesShopName(t *testing.T) {
	shop := testShop
	shop.Name = `<script>alert(1)</script>`
	out, err := HTML(HotDeal, LangEN, testOffer, shop, false)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "<script>alert(1)</script>") {
		t.Fatal("shop name was not escaped")
	}
}

func TestHTMLEmbedsPhoto(t *testing.T) {
	o := testOffer
	o.PhotoBytes = []byte("\x00\x00\x00\x1cftypavif\x00\x00\x00\x00avifmif1")
	o.PhotoType = "image/avif"
	out, err := HTML(HotDeal, LangEN, o, testShop, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "data:image/avif;base64,") {
		t.Fatal("photo should be embedded with its own content type")
	}
}

func TestSplitPrice(t *testing.T) {
	cases := map[float64][2]string{8: {"8", ".00"}, 24.99: {"24", ".99"}, 0.5: {"0", ".50"}}
	for in, want := range cases {
		i, d := splitPrice(in)
		if i != want[0] || d != want[1] {
			t.Errorf("splitPrice(%v) = %s%s, want %s%s", in, i, d, want[0], want[1])
		}
	}
}
