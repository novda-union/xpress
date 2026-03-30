package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/xpressgo/server/internal/config"
	"github.com/xpressgo/server/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type seedModifier struct {
	name  string
	price int64
}

type seedModifierGroup struct {
	name          string
	selectionType string
	required      bool
	mods          []seedModifier
}

type seedItem struct {
	category    string
	name        string
	description string
	price       int64
	modGroups   []seedModifierGroup
}

type seedStore struct {
	code        string
	slug        string
	name        string
	category    string
	description string
	address     string
	phone       string
	logoURL     string
	branches    []seedBranch
	staff       []seedStaff
	users       []seedUser
	menu        seedMenu
	orderPlan   seedOrderPlan
}

type seedBranch struct {
	code                string
	name                string
	address             string
	lat                 float64
	lng                 float64
	bannerImageURL      string
	telegramGroupChatID *int64
	isActive            bool
}

type seedStaff struct {
	branchCode string
	staffCode  string
	name       string
	role       string
	isActive   bool
}

type seedUser struct {
	telegramID int64
	phone      string
	firstName  string
	lastName   string
	username   string
}

type seedMenu struct {
	categories []seedCategory
}

type seedCategory struct {
	name  string
	items []seedCatalogItem
}

type seedCatalogItem struct {
	name          string
	description   string
	price         int64
	modGroups     []seedModifierGroup
	unavailableAt map[string]bool
}

type seedOrderPlan struct {
	totalOrders     int
	activePending   int
	activeAccepted  int
	activePreparing int
	activeReady     int
	branchWeights   []string
}

type seededStoreRefs struct {
	storeID   string
	branchIDs map[string]string
	userIDs   []string
	catalog   seededCatalog
}

type seededCatalog struct {
	itemsByBranch map[string][]seededMenuItem
}

type seededMenuItem struct {
	itemID     string
	name       string
	price      int64
	modifiers  []seededModifier
	branchCode string
}

type seededModifier struct {
	modifierID string
	name       string
	price      int64
}

type seededGeneratedOrder struct {
	userID          string
	storeID         string
	branchID        string
	status          string
	totalPrice      int64
	paymentMethod   string
	paymentStatus   string
	etaMinutes      int
	rejectionReason string
	createdAt       time.Time
	updatedAt       time.Time
	items           []seededGeneratedOrderItem
}

type seededGeneratedOrderItem struct {
	itemID    string
	itemName  string
	itemPrice int64
	quantity  int
	modifiers []seededGeneratedOrderModifier
}

type seededGeneratedOrderModifier struct {
	modifierID string
	name       string
	price      int64
}

func main() {
	cfg := config.Load()

	conn, err := database.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte("demo123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	stores := buildSeedStores()
	if len(stores) == 0 {
		log.Fatal("no seed stores configured")
	}

	if err := resetSeedUsers(ctx, conn); err != nil {
		log.Fatalf("failed to reset seed users: %v", err)
	}

	for _, store := range stores {
		refs, err := seedStoreWorld(ctx, conn, store, string(hash))
		if err != nil {
			log.Fatalf("failed to seed store world %s: %v", store.code, err)
		}

		if err := seedStoreOrders(ctx, conn, store, refs); err != nil {
			log.Fatalf("failed to seed orders for %s: %v", store.code, err)
		}

		log.Printf(
			"Store seeded: %s (%d branches, %d staff, %d users, %d orders)",
			store.name,
			len(store.branches),
			len(store.staff),
			len(refs.userIDs),
			store.orderPlan.totalOrders,
		)
	}

	log.Println("Seed completed successfully!")
}

func seedStoreWorld(ctx context.Context, conn *pgx.Conn, store seedStore, passwordHash string) (seededStoreRefs, error) {
	storeID, err := ensureStore(ctx, conn, store)
	if err != nil {
		return seededStoreRefs{}, err
	}

	refs := seededStoreRefs{
		storeID:   storeID,
		branchIDs: map[string]string{},
		userIDs:   []string{},
		catalog: seededCatalog{
			itemsByBranch: map[string][]seededMenuItem{},
		},
	}

	for _, branch := range store.branches {
		branchID, err := ensureBranch(ctx, conn, storeID, branch)
		if err != nil {
			return seededStoreRefs{}, err
		}
		refs.branchIDs[branch.code] = branchID
	}

	for _, staff := range store.staff {
		var branchID *string
		if staff.branchCode != "" {
			id := refs.branchIDs[staff.branchCode]
			branchID = &id
		}
		if err := ensureStaff(ctx, conn, storeID, branchID, staff.staffCode, staff.name, passwordHash, staff.role, staff.isActive); err != nil {
			return seededStoreRefs{}, err
		}
	}

	if err := deactivateStaleStaff(ctx, conn, storeID, store.staff); err != nil {
		return seededStoreRefs{}, err
	}

	for _, user := range store.users {
		userID, err := ensureUser(ctx, conn, user)
		if err != nil {
			return seededStoreRefs{}, err
		}
		refs.userIDs = append(refs.userIDs, userID)
	}

	for _, branch := range store.branches {
		branchID := refs.branchIDs[branch.code]
		items, err := seedBranchMenu(ctx, conn, storeID, branchID, store, branch)
		if err != nil {
			return seededStoreRefs{}, err
		}
		refs.catalog.itemsByBranch[branch.code] = items
	}

	return refs, nil
}

func seedStoreOrders(ctx context.Context, conn *pgx.Conn, store seedStore, refs seededStoreRefs) error {
	if len(refs.userIDs) == 0 {
		return nil
	}

	if err := deleteOrdersForSeedUsersByStore(ctx, conn, refs.storeID, refs.userIDs); err != nil {
		return err
	}

	statuses := buildOrderStatuses(store.orderPlan)
	now := time.Now().UTC()
	branchWeights := store.orderPlan.branchWeights
	if len(branchWeights) == 0 {
		for _, branch := range store.branches {
			branchWeights = append(branchWeights, branch.code)
		}
	}

	for i, status := range statuses {
		branchCode := branchWeights[i%len(branchWeights)]
		branchID := refs.branchIDs[branchCode]
		items := refs.catalog.itemsByBranch[branchCode]
		if len(items) == 0 {
			continue
		}

		orderItems, totalPrice := buildOrderItems(items, i)
		createdAt := buildOrderTimestamp(now, i, len(statuses))
		updatedAt := createdAt.Add(time.Duration(5+(i%90)) * time.Minute)
		if status == "pending" {
			updatedAt = createdAt
		}

		order := seededGeneratedOrder{
			userID:          refs.userIDs[i%len(refs.userIDs)],
			storeID:         refs.storeID,
			branchID:        branchID,
			status:          status,
			totalPrice:      totalPrice,
			paymentMethod:   "pay_at_pickup",
			paymentStatus:   paymentStatusForStatus(status),
			etaMinutes:      5 + (i%5)*5,
			rejectionReason: rejectionReasonForStatus(status, i),
			createdAt:       createdAt,
			updatedAt:       updatedAt,
			items:           orderItems,
		}

		if _, err := insertSeedOrder(ctx, conn, order); err != nil {
			return err
		}
	}

	return nil
}

func buildSeedStores() []seedStore {
	return []seedStore{
		buildDemoBarStore(),
		buildUrbanCoffeeStore(),
		buildStreetBurgerStore(),
	}
}

func buildDemoBarStore() seedStore {
	branchCodes := []string{"main", "downtown", "riverside", "chilonzor", "samarkand-darvoza", "airport-road"}
	return seedStore{
		code:        "demobar",
		slug:        "demo-bar",
		name:        "Demo Bar",
		category:    "bar",
		description: "The best cocktails in Tashkent",
		address:     "Amir Temur St 42, Tashkent",
		phone:       "+998901234567",
		branches: []seedBranch{
			{code: "main", name: "Demo Bar - Main", address: "Amir Temur St 42, Tashkent", lat: 41.2995, lng: 69.2401, isActive: true},
			{code: "downtown", name: "Demo Bar - Downtown", address: "Afrosiyob St 8, Tashkent", lat: 41.3057, lng: 69.2801, isActive: true},
			{code: "riverside", name: "Demo Bar - Riverside", address: "Kichik Halqa Yuli 14, Tashkent", lat: 41.2871, lng: 69.2684, isActive: true},
			{code: "chilonzor", name: "Demo Bar - Chilonzor", address: "Chilonzor 19 kvartal, Tashkent", lat: 41.2752, lng: 69.2038, isActive: true},
			{code: "samarkand-darvoza", name: "Demo Bar - Samarkand Darvoza", address: "Qatortol St 2, Tashkent", lat: 41.3164, lng: 69.2128, isActive: true},
			{code: "airport-road", name: "Demo Bar - Airport Road", address: "Kushbegi St 118, Tashkent", lat: 41.2578, lng: 69.2816, isActive: true},
		},
		staff: buildStoreStaff("Bar Director", branchCodes, 4, map[string]int{}),
		users: buildSeedUsers("demobaruser", "Demo", "Bar", 35, 2234500000),
		menu:  buildDemoBarMenu(),
		orderPlan: seedOrderPlan{
			totalOrders:     120,
			activePending:   4,
			activeAccepted:  5,
			activePreparing: 6,
			activeReady:     5,
			branchWeights:   []string{"main", "main", "downtown", "main", "riverside", "chilonzor", "samarkand-darvoza", "main", "airport-road"},
		},
	}
}

func buildUrbanCoffeeStore() seedStore {
	branchCodes := []string{"main", "yunusabad", "parkent"}
	return seedStore{
		code:        "urbancoffee",
		slug:        "urban-coffee",
		name:        "Urban Coffee",
		category:    "cafe",
		description: "Specialty coffee and breakfast",
		address:     "Buyuk Ipak Yuli 1, Tashkent",
		phone:       "+998901234568",
		branches: []seedBranch{
			{code: "main", name: "Urban Coffee - Main", address: "Buyuk Ipak Yuli 1, Tashkent", lat: 41.3142, lng: 69.2879, isActive: true},
			{code: "yunusabad", name: "Urban Coffee - Yunusabad", address: "Yunusabad 17, Tashkent", lat: 41.3647, lng: 69.2965, isActive: true},
			{code: "parkent", name: "Urban Coffee - Parkent", address: "Parkent St 12, Tashkent", lat: 41.2898, lng: 69.3212, isActive: true},
		},
		staff: buildStoreStaff("Coffee Director", branchCodes, 3, map[string]int{"parkent": 3}),
		users: buildSeedUsers("urbancoffeeuser", "Urban", "Coffee", 12, 2234600000),
		menu:  buildUrbanCoffeeMenu(),
		orderPlan: seedOrderPlan{
			totalOrders:     45,
			activePending:   1,
			activeAccepted:  1,
			activePreparing: 2,
			activeReady:     1,
			branchWeights:   []string{"main", "main", "yunusabad", "main", "parkent"},
		},
	}
}

func buildStreetBurgerStore() seedStore {
	branchCodes := []string{"main", "almazar", "sergeli"}
	return seedStore{
		code:        "streetburger",
		slug:        "street-burger",
		name:        "Street Burger",
		category:    "fastfood",
		description: "Burgers, chicken, and quick comfort food",
		address:     "Bodomzor 3, Tashkent",
		phone:       "+998901234569",
		branches: []seedBranch{
			{code: "main", name: "Street Burger - Main", address: "Bodomzor 3, Tashkent", lat: 41.3471, lng: 69.2872, isActive: true},
			{code: "almazar", name: "Street Burger - Almazar", address: "Almazar 8, Tashkent", lat: 41.3502, lng: 69.2198, isActive: true},
			{code: "sergeli", name: "Street Burger - Sergeli", address: "Sergeli 9, Tashkent", lat: 41.2259, lng: 69.2144, isActive: true},
		},
		staff: buildStoreStaff("Burger Director", branchCodes, 3, map[string]int{"sergeli": 2}),
		users: buildSeedUsers("streetburgeruser", "Street", "Burger", 12, 2234700000),
		menu:  buildStreetBurgerMenu(),
		orderPlan: seedOrderPlan{
			totalOrders:     45,
			activePending:   1,
			activeAccepted:  1,
			activePreparing: 2,
			activeReady:     1,
			branchWeights:   []string{"main", "main", "almazar", "main", "sergeli"},
		},
	}
}

func buildStoreStaff(directorName string, branchCodes []string, baristasPerBranch int, inactiveBaristas map[string]int) []seedStaff {
	staff := []seedStaff{
		{staffCode: "admin", name: directorName, role: "director", isActive: true},
	}

	for _, branchCode := range branchCodes {
		staff = append(staff, seedStaff{
			branchCode: branchCode,
			staffCode:  "manager-" + branchCode,
			name:       titleCase(branchCode) + " Manager",
			role:       "manager",
			isActive:   true,
		})

		for i := 1; i <= baristasPerBranch; i++ {
			isActive := true
			if inactiveBaristas[branchCode] == i {
				isActive = false
			}
			staff = append(staff, seedStaff{
				branchCode: branchCode,
				staffCode:  "barista-" + branchCode + "-" + itoa(i),
				name:       titleCase(branchCode) + " Barista " + itoa(i),
				role:       "barista",
				isActive:   isActive,
			})
		}
	}

	return staff
}

func buildSeedUsers(prefix, firstPrefix, lastPrefix string, count int, startTelegramID int64) []seedUser {
	users := make([]seedUser, 0, count)
	for i := 1; i <= count; i++ {
		suffix := itoa(i)
		users = append(users, seedUser{
			telegramID: startTelegramID + int64(i),
			phone:      "+99890" + leftPadNumber(i, 7),
			firstName:  firstPrefix,
			lastName:   lastPrefix + " User " + suffix,
			username:   prefix + suffix,
		})
	}
	return users
}

func buildDemoBarMenu() seedMenu {
	return seedMenu{
		categories: []seedCategory{
			{name: "Cocktails", items: []seedCatalogItem{
				{name: "Mojito", description: "Fresh lime, mint, rum, soda", price: 45000, modGroups: []seedModifierGroup{
					singleSelectGroup("Size", true, seedModifier{name: "Regular", price: 0}, seedModifier{name: "Large", price: 15000}),
					multiSelectGroup("Extras", seedModifier{name: "Extra mint", price: 3000}, seedModifier{name: "Extra lime", price: 3000}, seedModifier{name: "Double shot", price: 10000}),
				}, unavailableAt: map[string]bool{"airport-road": true}},
				{name: "Margarita", description: "Tequila, lime juice, triple sec", price: 50000, modGroups: []seedModifierGroup{
					singleSelectGroup("Size", true, seedModifier{name: "Regular", price: 0}, seedModifier{name: "Large", price: 15000}),
					singleSelectGroup("Salt Rim", false, seedModifier{name: "With salt", price: 0}, seedModifier{name: "No salt", price: 0}),
				}},
				{name: "Long Island", description: "Five spirits, cola, lemon", price: 55000, modGroups: []seedModifierGroup{
					singleSelectGroup("Size", true, seedModifier{name: "Regular", price: 0}, seedModifier{name: "Large", price: 20000}),
				}},
				{name: "Aperol Spritz", description: "Aperol, prosecco, soda", price: 48000},
				{name: "Negroni", description: "Gin, vermouth, campari", price: 52000},
			}},
			{name: "Beer", items: []seedCatalogItem{
				{name: "Sarbast Lager", description: "Local Uzbek lager, crisp and light", price: 25000, modGroups: []seedModifierGroup{
					singleSelectGroup("Size", true, seedModifier{name: "0.33L", price: 0}, seedModifier{name: "0.5L", price: 10000}),
				}},
				{name: "Pulsar IPA", description: "Hoppy craft IPA", price: 35000, modGroups: []seedModifierGroup{
					singleSelectGroup("Size", true, seedModifier{name: "0.33L", price: 0}, seedModifier{name: "0.5L", price: 12000}),
				}},
				{name: "Corona", description: "Mexican lager with lime", price: 40000},
				{name: "Wheat Ale", description: "Cloudy wheat beer", price: 32000},
				{name: "Pilsner", description: "Dry and refreshing lager", price: 28000},
			}},
			{name: "Wine", items: []seedCatalogItem{
				{name: "House White", description: "Crisp glass of white wine", price: 38000},
				{name: "House Red", description: "Smooth glass of red wine", price: 39000},
				{name: "Rose Spritz", description: "Rose with tonic and citrus", price: 41000},
				{name: "Sparkling Brut", description: "Light sparkling wine", price: 45000},
			}},
			{name: "Coffee", items: []seedCatalogItem{
				{name: "Espresso", description: "Strong single shot", price: 16000},
				{name: "Americano", description: "Espresso with hot water", price: 18000},
				{name: "Cappuccino", description: "Espresso with milk foam", price: 22000},
				{name: "Iced Latte", description: "Cold latte over ice", price: 26000},
			}},
			{name: "Snacks", items: []seedCatalogItem{
				{name: "Nachos", description: "Tortilla chips with cheese sauce and jalapenos", price: 35000, modGroups: []seedModifierGroup{
					multiSelectGroup("Toppings", seedModifier{name: "Extra cheese", price: 5000}, seedModifier{name: "Guacamole", price: 8000}, seedModifier{name: "Sour cream", price: 5000}),
				}},
				{name: "Chicken Wings", description: "Crispy wings with sauce", price: 42000, modGroups: []seedModifierGroup{
					singleSelectGroup("Sauce", true, seedModifier{name: "BBQ", price: 0}, seedModifier{name: "Buffalo", price: 0}, seedModifier{name: "Honey Mustard", price: 0}),
					singleSelectGroup("Portion", true, seedModifier{name: "6 pieces", price: 0}, seedModifier{name: "12 pieces", price: 25000}),
				}},
				{name: "French Fries", description: "Crispy golden fries", price: 20000, modGroups: []seedModifierGroup{
					singleSelectGroup("Dip", false, seedModifier{name: "Ketchup", price: 0}, seedModifier{name: "Mayo", price: 0}, seedModifier{name: "Cheese sauce", price: 3000}),
				}},
				{name: "Onion Rings", description: "Crunchy onion rings", price: 26000},
			}},
			{name: "Sharing Plates", items: []seedCatalogItem{
				{name: "Mixed Platter", description: "Wings, fries, rings, and dips", price: 72000},
				{name: "Cheese Board", description: "Selection of cheeses and nuts", price: 68000, unavailableAt: map[string]bool{"chilonzor": true}},
				{name: "Loaded Fries", description: "Fries with cheese, jalapenos, and sauce", price: 44000},
				{name: "Mini Sliders", description: "Three beef sliders", price: 52000},
			}},
		},
	}
}

func buildUrbanCoffeeMenu() seedMenu {
	return seedMenu{
		categories: []seedCategory{
			{name: "Espresso Bar", items: []seedCatalogItem{
				{name: "Espresso", description: "Strong single shot", price: 12000},
				{name: "Double Espresso", description: "Double shot", price: 16000},
				{name: "Americano", description: "Espresso and hot water", price: 14000},
				{name: "Flat White", description: "Velvety milk coffee", price: 18000},
				{name: "Cappuccino", description: "Espresso with milk foam", price: 19000},
			}},
			{name: "Signature Drinks", items: []seedCatalogItem{
				{name: "Iced Latte", description: "Cold latte over ice", price: 22000},
				{name: "Caramel Macchiato", description: "Espresso with caramel syrup", price: 24000},
				{name: "Vanilla Raf", description: "Creamy vanilla coffee", price: 26000},
				{name: "Mocha", description: "Coffee with chocolate", price: 23000},
				{name: "Cold Brew", description: "Slow-steeped iced coffee", price: 25000, unavailableAt: map[string]bool{"parkent": true}},
			}},
			{name: "Tea", items: []seedCatalogItem{
				{name: "English Breakfast", description: "Classic black tea", price: 15000},
				{name: "Jasmine Green Tea", description: "Light floral green tea", price: 16000},
				{name: "Chamomile", description: "Calming herbal tea", price: 16000},
				{name: "Lemon Ginger Tea", description: "Citrus and ginger infusion", price: 18000},
			}},
			{name: "Pastries", items: []seedCatalogItem{
				{name: "Butter Croissant", description: "Flaky butter croissant", price: 16000},
				{name: "Chocolate Croissant", description: "Pastry with dark chocolate", price: 19000},
				{name: "Cinnamon Roll", description: "Soft cinnamon pastry", price: 22000},
				{name: "Blueberry Muffin", description: "Muffin with fresh berries", price: 20000},
			}},
			{name: "Breakfast", items: []seedCatalogItem{
				{name: "Avocado Toast", description: "Sourdough with avocado and herbs", price: 36000},
				{name: "Egg Sandwich", description: "Scrambled eggs and cheese", price: 32000},
				{name: "Granola Bowl", description: "Yogurt with granola and fruit", price: 28000},
				{name: "Pancake Stack", description: "Pancakes with syrup and berries", price: 34000},
			}},
		},
	}
}

func buildStreetBurgerMenu() seedMenu {
	return seedMenu{
		categories: []seedCategory{
			{name: "Burgers", items: []seedCatalogItem{
				{name: "Classic Burger", description: "Beef patty, lettuce, tomato, pickles", price: 45000, modGroups: []seedModifierGroup{
					singleSelectGroup("Doneness", false, seedModifier{name: "Medium", price: 0}, seedModifier{name: "Well Done", price: 0}),
				}},
				{name: "Cheeseburger", description: "Burger with cheddar", price: 48000},
				{name: "Double Burger", description: "Two patties and cheese", price: 55000},
				{name: "BBQ Burger", description: "BBQ sauce and onion rings", price: 50000},
				{name: "Mushroom Burger", description: "Mushrooms and creamy sauce", price: 52000},
			}},
			{name: "Chicken", items: []seedCatalogItem{
				{name: "Crispy Chicken Burger", description: "Fried chicken with slaw", price: 43000},
				{name: "Spicy Chicken Wrap", description: "Wrap with spicy chicken strips", price: 39000},
				{name: "Chicken Tenders", description: "Tender strips with dip", price: 35000},
				{name: "Buffalo Chicken Burger", description: "Chicken burger with buffalo sauce", price: 47000},
			}},
			{name: "Sides", items: []seedCatalogItem{
				{name: "Fries", description: "Classic salted fries", price: 18000},
				{name: "Loaded Fries", description: "Fries with cheese and sauce", price: 28000},
				{name: "Onion Rings", description: "Crispy onion rings", price: 22000},
				{name: "Coleslaw", description: "Fresh slaw side", price: 12000},
			}},
			{name: "Combos", items: []seedCatalogItem{
				{name: "Burger Combo", description: "Burger, fries, and drink", price: 62000},
				{name: "Chicken Combo", description: "Chicken burger, fries, and drink", price: 60000},
				{name: "Double Combo", description: "Double burger set with fries", price: 72000, unavailableAt: map[string]bool{"sergeli": true}},
			}},
			{name: "Drinks", items: []seedCatalogItem{
				{name: "Cola", description: "Chilled cola", price: 12000},
				{name: "Lemonade", description: "Sparkling lemonade", price: 14000},
				{name: "Iced Tea", description: "Sweet iced tea", price: 13000},
				{name: "Milkshake", description: "Vanilla milkshake", price: 22000},
			}},
		},
	}
}

func singleSelectGroup(name string, required bool, mods ...seedModifier) seedModifierGroup {
	return seedModifierGroup{
		name:          name,
		selectionType: "single",
		required:      required,
		mods:          mods,
	}
}

func multiSelectGroup(name string, mods ...seedModifier) seedModifierGroup {
	return seedModifierGroup{
		name:          name,
		selectionType: "multiple",
		required:      false,
		mods:          mods,
	}
}

func ensureStore(ctx context.Context, conn *pgx.Conn, store seedStore) (string, error) {
	var id string
	err := conn.QueryRow(ctx, `
		INSERT INTO stores (
			name, code, slug, category, description, address, phone, logo_url,
			telegram_group_chat_id, subscription_tier, commission_rate, is_active
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			NULL, 'free', 5.00, true
		)
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			slug = EXCLUDED.slug,
			category = EXCLUDED.category,
			description = EXCLUDED.description,
			address = EXCLUDED.address,
			phone = EXCLUDED.phone,
			logo_url = EXCLUDED.logo_url,
			telegram_group_chat_id = EXCLUDED.telegram_group_chat_id,
			subscription_tier = EXCLUDED.subscription_tier,
			commission_rate = EXCLUDED.commission_rate,
			is_active = EXCLUDED.is_active,
			updated_at = NOW()
		RETURNING id
	`, store.name, store.code, store.slug, store.category, store.description, store.address, store.phone, store.logoURL).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func ensureBranch(ctx context.Context, conn *pgx.Conn, storeID string, branch seedBranch) (string, error) {
	var id string
	err := conn.QueryRow(ctx, `
		INSERT INTO branches (
			store_id, name, address, lat, lng, banner_image_url, telegram_group_chat_id, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (store_id, name) DO UPDATE SET
			address = EXCLUDED.address,
			lat = EXCLUDED.lat,
			lng = EXCLUDED.lng,
			banner_image_url = EXCLUDED.banner_image_url,
			telegram_group_chat_id = EXCLUDED.telegram_group_chat_id,
			is_active = EXCLUDED.is_active,
			updated_at = NOW()
		RETURNING id
	`, storeID, branch.name, branch.address, branch.lat, branch.lng, branch.bannerImageURL, branch.telegramGroupChatID, branch.isActive).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func ensureUser(ctx context.Context, conn *pgx.Conn, user seedUser) (string, error) {
	var id string
	err := conn.QueryRow(ctx, `
		INSERT INTO users (telegram_id, phone, first_name, last_name, username)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (telegram_id) DO UPDATE SET
			phone = EXCLUDED.phone,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			username = EXCLUDED.username
		RETURNING id
	`, user.telegramID, user.phone, user.firstName, user.lastName, user.username).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func ensureStaff(ctx context.Context, conn *pgx.Conn, storeID string, branchID *string, staffCode, name, passwordHash, role string, isActive bool) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO store_staff (store_id, branch_id, staff_code, name, password_hash, role, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (store_id, staff_code) DO UPDATE SET
			branch_id = EXCLUDED.branch_id,
			name = EXCLUDED.name,
			password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role,
			is_active = EXCLUDED.is_active
	`, storeID, branchID, staffCode, name, passwordHash, role, isActive)
	return err
}

func ensureCategory(ctx context.Context, conn *pgx.Conn, storeID, branchID, name string, sortOrder int) (string, error) {
	var id string
	err := conn.QueryRow(ctx, `
		SELECT id
		FROM categories
		WHERE store_id = $1 AND branch_id = $2 AND name = $3
	`, storeID, branchID, name).Scan(&id)
	if err == nil {
		_, err = conn.Exec(ctx, `
			UPDATE categories
			SET sort_order = $2, is_active = true
			WHERE id = $1
		`, id, sortOrder)
		return id, err
	}
	if err != pgx.ErrNoRows {
		return "", err
	}

	err = conn.QueryRow(ctx, `
		INSERT INTO categories (store_id, branch_id, name, sort_order, is_active)
		VALUES ($1, $2, $3, $4, true)
		RETURNING id
	`, storeID, branchID, name, sortOrder).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func ensureItem(ctx context.Context, conn *pgx.Conn, storeID, branchID, categoryID string, item seedItem, sortOrder int) (string, error) {
	var id string
	err := conn.QueryRow(ctx, `
		SELECT id
		FROM items
		WHERE store_id = $1 AND branch_id = $2 AND name = $3
	`, storeID, branchID, item.name).Scan(&id)
	if err == nil {
		_, err = conn.Exec(ctx, `
			UPDATE items
			SET category_id = $2, description = $3, base_price = $4, image_url = $5, is_available = true, sort_order = $6
			WHERE id = $1
		`, id, categoryID, item.description, item.price, "", sortOrder)
		return id, err
	}
	if err != pgx.ErrNoRows {
		return "", err
	}

	err = conn.QueryRow(ctx, `
		INSERT INTO items (category_id, store_id, branch_id, name, description, base_price, image_url, is_available, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true, $8)
		RETURNING id
	`, categoryID, storeID, branchID, item.name, item.description, item.price, "", sortOrder).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func ensureModifierGroup(ctx context.Context, conn *pgx.Conn, storeID, branchID, itemID string, group seedModifierGroup, sortOrder int) (string, error) {
	var id string
	err := conn.QueryRow(ctx, `
		SELECT id
		FROM modifier_groups
		WHERE store_id = $1 AND branch_id = $2 AND item_id = $3 AND name = $4
	`, storeID, branchID, itemID, group.name).Scan(&id)
	if err == nil {
		minSel := 0
		maxSel := 1
		if group.required {
			minSel = 1
		}
		if group.selectionType == "multiple" {
			maxSel = len(group.mods)
		}
		_, err = conn.Exec(ctx, `
			UPDATE modifier_groups
			SET selection_type = $2, is_required = $3, min_selections = $4, max_selections = $5, sort_order = $6
			WHERE id = $1
		`, id, group.selectionType, group.required, minSel, maxSel, sortOrder)
		return id, err
	}
	if err != pgx.ErrNoRows {
		return "", err
	}

	minSel := 0
	maxSel := 1
	if group.required {
		minSel = 1
	}
	if group.selectionType == "multiple" {
		maxSel = len(group.mods)
	}
	err = conn.QueryRow(ctx, `
		INSERT INTO modifier_groups (
			item_id, store_id, branch_id, name, selection_type, is_required, min_selections, max_selections, sort_order
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, itemID, storeID, branchID, group.name, group.selectionType, group.required, minSel, maxSel, sortOrder).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func ensureModifier(ctx context.Context, conn *pgx.Conn, storeID, branchID, groupID string, mod seedModifier, sortOrder int) (string, error) {
	var id string
	err := conn.QueryRow(ctx, `
		SELECT id
		FROM modifiers
		WHERE store_id = $1 AND branch_id = $2 AND modifier_group_id = $3 AND name = $4
	`, storeID, branchID, groupID, mod.name).Scan(&id)
	if err == nil {
		_, err = conn.Exec(ctx, `
			UPDATE modifiers
			SET price_adjustment = $2, is_available = true, sort_order = $3
			WHERE id = $1
		`, id, mod.price, sortOrder)
		return id, err
	}
	if err != pgx.ErrNoRows {
		return "", err
	}

	err = conn.QueryRow(ctx, `
		INSERT INTO modifiers (modifier_group_id, store_id, branch_id, name, price_adjustment, is_available, sort_order)
		VALUES ($1, $2, $3, $4, $5, true, $6)
		RETURNING id
	`, groupID, storeID, branchID, mod.name, mod.price, sortOrder).Scan(&id)
	return id, err
}

func seedBranchMenu(ctx context.Context, conn *pgx.Conn, storeID, branchID string, store seedStore, branch seedBranch) ([]seededMenuItem, error) {
	items := []seededMenuItem{}

	for categoryIndex, category := range store.menu.categories {
		categoryID, err := ensureCategory(ctx, conn, storeID, branchID, category.name, categoryIndex)
		if err != nil {
			return nil, err
		}

		for itemIndex, catalogItem := range category.items {
			itemID, err := ensureItem(ctx, conn, storeID, branchID, categoryID, seedItem{
				category:    category.name,
				name:        catalogItem.name,
				description: catalogItem.description,
				price:       catalogItem.price,
				modGroups:   catalogItem.modGroups,
			}, itemIndex)
			if err != nil {
				return nil, err
			}

			if err := setItemAvailability(ctx, conn, itemID, !catalogItem.unavailableAt[branch.code]); err != nil {
				return nil, err
			}

			menuItem := seededMenuItem{
				itemID:     itemID,
				name:       catalogItem.name,
				price:      catalogItem.price,
				modifiers:  []seededModifier{},
				branchCode: branch.code,
			}

			for groupIndex, group := range catalogItem.modGroups {
				groupID, err := ensureModifierGroup(ctx, conn, storeID, branchID, itemID, group, groupIndex)
				if err != nil {
					return nil, err
				}

				for modIndex, mod := range group.mods {
					modifierID, err := ensureModifier(ctx, conn, storeID, branchID, groupID, mod, modIndex)
					if err != nil {
						return nil, err
					}
					menuItem.modifiers = append(menuItem.modifiers, seededModifier{
						modifierID: modifierID,
						name:       mod.name,
						price:      mod.price,
					})
				}
			}

			items = append(items, menuItem)
		}
	}

	return items, nil
}

func setItemAvailability(ctx context.Context, conn *pgx.Conn, itemID string, isAvailable bool) error {
	_, err := conn.Exec(ctx, `
		UPDATE items
		SET is_available = $2
		WHERE id = $1
	`, itemID, isAvailable)
	return err
}

func deleteOrdersForSeedUsersByStore(ctx context.Context, conn *pgx.Conn, storeID string, userIDs []string) error {
	_, err := conn.Exec(ctx, `
		DELETE FROM orders
		WHERE store_id = $1
		  AND user_id = ANY($2)
	`, storeID, userIDs)
	return err
}

func resetSeedUsers(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `
		DELETE FROM orders
		WHERE user_id IN (
			SELECT id FROM users
			WHERE username = 'demouser'
			   OR username LIKE 'demobaruser%'
			   OR username LIKE 'urbancoffeeuser%'
			   OR username LIKE 'streetburgeruser%'
		)
	`)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, `
		DELETE FROM users
		WHERE username = 'demouser'
		   OR username LIKE 'demobaruser%'
		   OR username LIKE 'urbancoffeeuser%'
		   OR username LIKE 'streetburgeruser%'
	`)
	return err
}

func deactivateStaleStaff(ctx context.Context, conn *pgx.Conn, storeID string, staff []seedStaff) error {
	validCodes := make([]string, 0, len(staff))
	for _, entry := range staff {
		validCodes = append(validCodes, entry.staffCode)
	}

	_, err := conn.Exec(ctx, `
		UPDATE store_staff
		SET is_active = false
		WHERE store_id = $1
		  AND NOT (staff_code = ANY($2))
	`, storeID, validCodes)
	return err
}

func insertSeedOrder(ctx context.Context, conn *pgx.Conn, order seededGeneratedOrder) (string, error) {
	var orderID string
	err := conn.QueryRow(ctx, `
		INSERT INTO orders (
			user_id, store_id, branch_id, status, total_price, payment_method,
			payment_status, eta_minutes, rejection_reason, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, order.userID, order.storeID, order.branchID, order.status, order.totalPrice, order.paymentMethod, order.paymentStatus, order.etaMinutes, order.rejectionReason, order.createdAt, order.updatedAt).Scan(&orderID)
	if err != nil {
		return "", err
	}

	for _, item := range order.items {
		orderItemID, err := insertSeedOrderItem(ctx, conn, orderID, item)
		if err != nil {
			return "", err
		}

		for _, mod := range item.modifiers {
			if err := insertSeedOrderModifier(ctx, conn, orderItemID, mod); err != nil {
				return "", err
			}
		}
	}

	return orderID, nil
}

func insertSeedOrderItem(ctx context.Context, conn *pgx.Conn, orderID string, item seededGeneratedOrderItem) (string, error) {
	var orderItemID string
	err := conn.QueryRow(ctx, `
		INSERT INTO order_items (order_id, item_id, item_name, item_price, quantity)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, orderID, item.itemID, item.itemName, item.itemPrice, item.quantity).Scan(&orderItemID)
	return orderItemID, err
}

func insertSeedOrderModifier(ctx context.Context, conn *pgx.Conn, orderItemID string, mod seededGeneratedOrderModifier) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO order_item_modifiers (order_item_id, modifier_id, modifier_name, price_adjustment)
		VALUES ($1, $2, $3, $4)
	`, orderItemID, mod.modifierID, mod.name, mod.price)
	return err
}

func buildOrderStatuses(plan seedOrderPlan) []string {
	statuses := make([]string, 0, plan.totalOrders)

	for i := 0; i < plan.activePending; i++ {
		statuses = append(statuses, "pending")
	}
	for i := 0; i < plan.activeAccepted; i++ {
		statuses = append(statuses, "accepted")
	}
	for i := 0; i < plan.activePreparing; i++ {
		statuses = append(statuses, "preparing")
	}
	for i := 0; i < plan.activeReady; i++ {
		statuses = append(statuses, "ready")
	}

	for len(statuses) < plan.totalOrders {
		switch len(statuses) % 10 {
		case 0, 1, 2, 3, 4, 5, 6:
			statuses = append(statuses, "picked_up")
		case 7, 8:
			statuses = append(statuses, "cancelled")
		default:
			statuses = append(statuses, "rejected")
		}
	}

	return statuses
}

func buildOrderTimestamp(now time.Time, index, total int) time.Time {
	activeWindow := total / 3
	if index < activeWindow {
		return now.Add(-time.Duration(1+(index*3)%42) * time.Hour)
	}

	daysAgo := 2 + (index % 12)
	minutes := (index * 37) % (24 * 60)
	return now.Add(-time.Duration(daysAgo)*24*time.Hour - time.Duration(minutes)*time.Minute)
}

func buildOrderItems(items []seededMenuItem, orderIndex int) ([]seededGeneratedOrderItem, int64) {
	count := 1 + (orderIndex % 3)
	if orderIndex%10 == 0 {
		count = 4
	}

	result := make([]seededGeneratedOrderItem, 0, count)
	var total int64
	for i := 0; i < count; i++ {
		menuItem := items[(orderIndex+i)%len(items)]
		quantity := 1
		if (orderIndex+i)%5 == 0 {
			quantity = 2
		}

		generated := seededGeneratedOrderItem{
			itemID:    menuItem.itemID,
			itemName:  menuItem.name,
			itemPrice: menuItem.price,
			quantity:  quantity,
			modifiers: []seededGeneratedOrderModifier{},
		}

		lineTotal := menuItem.price * int64(quantity)
		if len(menuItem.modifiers) > 0 && (orderIndex+i)%2 == 0 {
			mod := menuItem.modifiers[(orderIndex+i)%len(menuItem.modifiers)]
			generated.modifiers = append(generated.modifiers, seededGeneratedOrderModifier(mod))
			lineTotal += mod.price * int64(quantity)
		}

		total += lineTotal
		result = append(result, generated)
	}

	return result, total
}

func paymentStatusForStatus(status string) string {
	switch status {
	case "picked_up":
		return "paid"
	case "rejected":
		return "failed"
	case "cancelled":
		return "cancelled"
	default:
		return "pending"
	}
}

func rejectionReasonForStatus(status string, index int) string {
	if status != "rejected" {
		return ""
	}
	if index%2 == 0 {
		return "Kitchen capacity reached"
	}
	return "Item temporarily unavailable"
}

func titleCase(code string) string {
	runes := []rune(code)
	if len(runes) == 0 {
		return ""
	}
	for i := range runes {
		if i == 0 || runes[i-1] == '-' {
			if runes[i] >= 'a' && runes[i] <= 'z' {
				runes[i] = runes[i] - 32
			}
		} else if runes[i] == '-' {
			runes[i] = ' '
		}
	}
	return string(runes)
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}

	buf := [20]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + (v % 10))
		v /= 10
	}
	return string(buf[i:])
}

func leftPadNumber(v, width int) string {
	s := itoa(v)
	for len(s) < width {
		s = "0" + s
	}
	return s
}
