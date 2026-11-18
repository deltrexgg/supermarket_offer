// Package seed fills an empty database with Dubai demo data and creates the demo
// logins in Supabase Auth. Offer dates are relative to today (Dubai), so the demo
// always has active, upcoming, ending-today and expired offers.
package seed

import (
	"context"
	"fmt"
	"log"

	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/dubai"
	"github.com/deltrexgg/supermarket_offer/api/internal/models"
)

const DemoPassword = "Offers@12345"

type market struct {
	name, slug, phone, email string
	branches                 [2][3]string // name, address, area
}

var markets = []market{
	{"Al Noor Fresh Mart", "al-noor-fresh-mart", "+971 4 221 0001", "noor@offers.test",
		[2][3]string{{"Deira City Centre", "Al Rigga Rd, Deira", "Deira"}, {"Bur Dubai Souk", "Meena Bazaar, Bur Dubai", "Bur Dubai"}}},
	{"Gulf Basket", "gulf-basket", "+971 4 221 0002", "gulf@offers.test",
		[2][3]string{{"Marina Walk", "Marina Walk, Dubai Marina", "Marina"}, {"Al Barsha Mall", "Al Barsha 1, Dubai", "Al Barsha"}}},
	{"Spice Garden Hypermarket", "spice-garden", "+971 4 221 0003", "spice@offers.test",
		[2][3]string{{"Deira Main", "Naif Rd, Deira", "Deira"}, {"Barsha South", "Barsha South 2, Dubai", "Al Barsha"}}},
}

var areas = []string{"Deira", "Bur Dubai", "Marina", "Al Barsha"}
var categories = []string{"Groceries", "Dairy", "Beverages", "Fruits & Vegetables", "Meat & Fish", "Household", "Bakery"}

type item struct {
	name, ar, brand, unit, cat string
	regular, offer             float64
}

// 10 items per supermarket; the day offsets below give each a different status.
var items = [][]item{
	{
		{"Basmati Rice", "أرز بسمتي", "India Gate", "5 kg", "Groceries", 39.95, 29.95},
		{"Full Cream Milk", "حليب كامل الدسم", "Almarai", "2 L", "Dairy", 11.50, 8.95},
		{"Orange Juice", "عصير برتقال", "Rubicon", "1 L", "Beverages", 7.25, 5.50},
		{"Red Apples", "تفاح أحمر", "", "1 kg", "Fruits & Vegetables", 9.95, 6.95},
		{"Chicken Breast", "صدر دجاج", "Al Islami", "1 kg", "Meat & Fish", 24.00, 18.50},
		{"Dish Soap", "سائل غسيل الصحون", "Fairy", "750 ml", "Household", 14.95, 10.95},
		{"Arabic Bread", "خبز عربي", "", "6 pcs", "Bakery", 3.50, 2.50},
		{"Cooking Oil", "زيت طبخ", "Sunflower Gold", "3 L", "Groceries", 32.00, 24.99},
		{"Greek Yogurt", "زبادي يوناني", "Danone", "500 g", "Dairy", 8.50, 6.25},
		{"Mineral Water", "مياه معدنية", "Masafi", "12 x 500 ml", "Beverages", 12.00, 9.00},
	},
	{
		{"Jasmine Rice", "أرز ياسمين", "Tilda", "5 kg", "Groceries", 45.00, 34.50},
		{"Labneh", "لبنة", "Puck", "400 g", "Dairy", 7.95, 5.95},
		{"Cola", "كولا", "Pepsi", "6 x 330 ml", "Beverages", 14.00, 10.50},
		{"Bananas", "موز", "", "1 kg", "Fruits & Vegetables", 5.95, 3.95},
		{"Salmon Fillet", "فيليه سلمون", "Norwegian", "500 g", "Meat & Fish", 49.00, 38.00},
		{"Laundry Liquid", "سائل غسيل", "Ariel", "3 L", "Household", 42.00, 31.50},
		{"Croissant", "كرواسان", "", "4 pcs", "Bakery", 12.00, 8.99},
		{"Pasta", "معكرونة", "Barilla", "500 g", "Groceries", 6.50, 4.50},
		{"Cheddar Cheese", "جبن شيدر", "Kraft", "400 g", "Dairy", 18.50, 14.00},
		{"Green Tea", "شاي أخضر", "Lipton", "100 bags", "Beverages", 19.95, 14.95},
	},
	{
		{"Atta Flour", "دقيق أتا", "Aashirvaad", "10 kg", "Groceries", 38.00, 29.00},
		{"Fresh Cream", "كريمة طازجة", "Puck", "250 ml", "Dairy", 5.50, 4.25},
		{"Mango Juice", "عصير مانجو", "Al Rawabi", "1 L", "Beverages", 6.95, 4.95},
		{"Tomatoes", "طماطم", "", "1 kg", "Fruits & Vegetables", 4.95, 2.95},
		{"Lamb Chops", "ريش لحم ضأن", "Local", "1 kg", "Meat & Fish", 79.00, 64.00},
		{"Paper Towels", "مناديل مطبخ", "Fine", "6 rolls", "Household", 21.00, 15.95},
		{"Whole Wheat Toast", "توست قمح كامل", "Modern Bakery", "600 g", "Bakery", 7.00, 5.25},
		{"Red Lentils", "عدس أحمر", "Tata", "2 kg", "Groceries", 16.50, 12.50},
		{"Butter", "زبدة", "Lurpak", "200 g", "Dairy", 13.50, 10.95},
		{"Energy Drink", "مشروب طاقة", "Red Bull", "4 x 250 ml", "Beverages", 22.00, 17.50},
	},
}

// {from, to} offsets in days from today (Dubai) for the 10 items of each market:
// active, ends today, upcoming, expired, long-running.
var windows = [10][2]int{
	{-3, 4}, {-6, 0}, {2, 9}, {-12, -3}, {-1, 20},
	{0, 6}, {-5, 0}, {1, 7}, {-20, -8}, {-2, 12},
}

// Run seeds the database when it has no supermarkets, then makes sure the demo
// logins exist in Supabase Auth and are linked in the users table.
func Run(ctx context.Context, db *gorm.DB, supabaseURL, secretKey string) error {
	if supabaseURL == "" || secretKey == "" {
		return fmt.Errorf("SEED needs SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY")
	}
	var n int64
	if err := db.WithContext(ctx).Model(&models.Supermarket{}).Count(&n).Error; err != nil {
		return err
	}
	if n == 0 {
		if err := seedData(ctx, db); err != nil {
			return err
		}
		log.Printf("seed: inserted %d supermarkets, %d branches, %d offers", len(markets), len(markets)*2, len(markets)*10)
	} else {
		log.Printf("seed: data already present, skipping inserts")
	}
	return seedLogins(ctx, db, supabaseURL, secretKey)
}

func seedData(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		areaID := map[string]uint{}
		for _, a := range areas {
			row := models.Area{Name: a}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			areaID[a] = row.ID
		}
		catID := map[string]uint{}
		for _, c := range categories {
			row := models.Category{Name: c}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			catID[c] = row.ID
		}
		for mi, m := range markets {
			phone := m.phone
			sm := models.Supermarket{Name: m.name, Slug: m.slug, Phone: &phone, Status: "approved"}
			if err := tx.Create(&sm).Error; err != nil {
				return err
			}
			var branchIDs []uint
			for _, b := range m.branches {
				br := models.Branch{SupermarketID: sm.ID, AreaID: areaID[b[2]], Name: b[0], Address: b[1]}
				if err := tx.Create(&br).Error; err != nil {
					return err
				}
				branchIDs = append(branchIDs, br.ID)
			}
			for i, it := range items[mi] {
				o := models.Offer{
					SupermarketID: sm.ID, CategoryID: catID[it.cat],
					ItemName: it.name, Unit: it.unit,
					RegularPrice: it.regular, OfferPrice: it.offer,
					ValidFrom: dubai.DateOffset(windows[i][0]), ValidTo: dubai.DateOffset(windows[i][1]),
				}
				ar := it.ar
				o.ItemNameAr = &ar
				if it.brand != "" {
					b := it.brand
					o.Brand = &b
				}
				if err := tx.Create(&o).Error; err != nil {
					return err
				}
				// Every offer is on branch 1; every third one is on both branches.
				links := branchIDs[:1]
				if i%3 == 0 {
					links = branchIDs
				}
				for _, bid := range links {
					if err := tx.Create(&models.OfferBranch{OfferID: o.ID, BranchID: bid, SupermarketID: sm.ID}).Error; err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

func seedLogins(ctx context.Context, db *gorm.DB, url, key string) error {
	ac := newAdminClient(url, key)

	link := func(email, role string, smID *uint) error {
		var existing models.User
		err := db.WithContext(ctx).Where("email = ?", email).First(&existing).Error
		if err == nil {
			return nil
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		authID, err := ac.ensureUser(ctx, email, DemoPassword)
		if err != nil {
			return err
		}
		return db.WithContext(ctx).Create(&models.User{AuthUserID: authID, Email: email, Role: role, SupermarketID: smID}).Error
	}

	for _, m := range markets {
		var sm models.Supermarket
		if err := db.WithContext(ctx).Where("slug = ?", m.slug).First(&sm).Error; err != nil {
			return fmt.Errorf("supermarket %s: %w", m.slug, err)
		}
		id := sm.ID
		if err := link(m.email, "supermarket", &id); err != nil {
			return fmt.Errorf("login %s: %w", m.email, err)
		}
	}
	log.Printf("seed: demo logins ready (password %s)", DemoPassword)
	return nil
}
