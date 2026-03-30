package main

import (
	"context"
	"fmt"
	"log"
	"sort"
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

type seededStoreRefs struct {
	storeID   string
	branchIDs map[string]string
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
	totalOrders int
}

type seedGeneratedOrder struct {
	userID          string
	storeID         string
	branchID        string
	status          string
	paymentMethod   string
	paymentStatus   string
	etaMinutes      int
	totalPrice      int64
	rejectionReason string
	createdAt       time.Time
	items           []seedGeneratedOrderItem
}

type seedGeneratedOrderItem struct {
	itemID    *string
	itemName  string
	itemPrice int64
	quantity  int
	modifiers []seedGeneratedOrderModifier
}

type seedGeneratedOrderModifier struct {
	modifierID      *string
	modifierName    string
	priceAdjustment int64
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
	for _, store := range stores {
		storeID, err := ensureStore(ctx, conn, store)
		if err != nil {
			log.Fatalf("failed to ensure store %s: %v", store.code, err)
		}
		log.Printf("Store ready: %s (%s)", store.name, storeID)

		refs := seededStoreRefs{
			storeID:   storeID,
			branchIDs: map[string]string{},
		}

		for _, branch := range store.branches {
			branchID, err := ensureBranch(ctx, conn, storeID, branch)
			if err != nil {
				log.Fatalf("failed to ensure branch %s/%s: %v", store.code, branch.code, err)
			}
			log.Printf("Branch ready: %s (%s)", branch.name, branchID)
			refs.branchIDs[branch.code] = branchID
		}

		userIDs := make([]string, 0, len(store.users))
		for _, user := range store.users {
			userID, err := ensureUser(ctx, conn, user)
			if err != nil {
				log.Fatalf("failed to ensure user %d for store %s: %v", user.telegramID, store.code, err)
			}
			userIDs = append(userIDs, userID)
		}

		for _, staff := range store.staff {
			var branchID *string
			if staff.branchCode != "" {
				id, ok := refs.branchIDs[staff.branchCode]
				if !ok {
					log.Fatalf("failed to resolve branch %s for staff %s/%s", staff.branchCode, store.code, staff.staffCode)
				}
				branchID = &id
			}

			if err := ensureStaff(ctx, conn, storeID, branchID, staff.staffCode, staff.name, string(hash), staff.role, staff.isActive); err != nil {
				log.Fatalf("failed to ensure staff %s/%s: %v", store.code, staff.staffCode, err)
			}
		}

		for _, branch := range store.branches {
			branchID, ok := refs.branchIDs[branch.code]
			if !ok {
				log.Fatalf("failed to resolve branch %s for menu seeding in store %s", branch.code, store.code)
			}
			if err := seedBranchMenu(ctx, conn, storeID, branchID, store, branch); err != nil {
				log.Fatalf("failed to seed menu for %s/%s: %v", store.code, branch.code, err)
			}
		}

		if err := replaceSeedOrders(ctx, conn, store, refs, userIDs); err != nil {
			log.Fatalf("failed to seed orders for %s: %v", store.code, err)
		}
	}

	log.Println("Seed completed successfully!")
}

func buildSeedStores() []seedStore {
	return []seedStore{
		buildDemoBarStore(),
		buildUrbanCoffeeStore(),
		buildStreetBurgerStore(),
	}
}

func buildDemoBarStore() seedStore {
	return seedStore{
		code:        "demobar",
		slug:        "demo-bar",
		name:        "Demo Bar",
		category:    "bar",
		description: "The best cocktails in Tashkent",
		address:     "Amir Temur St 42, Tashkent",
		phone:       "+998901234567",
		logoURL:     "",
		users: []seedUser{
			{telegramID: 123456789, phone: "+998901111111", firstName: "Demo", lastName: "User", username: "demouser"},
			{telegramID: 123456790, phone: "+998901111112", firstName: "Aisha", lastName: "Rahimova", username: "aisha_r"},
			{telegramID: 123456791, phone: "+998901111113", firstName: "Aziz", lastName: "Karimov", username: "aziz_k"},
			{telegramID: 123456792, phone: "+998901111114", firstName: "Dilnoza", lastName: "Saidova", username: "dilnoza_s"},
			{telegramID: 123456793, phone: "+998901111115", firstName: "Rustam", lastName: "Yusupov", username: "rustam_y"},
			{telegramID: 123456794, phone: "+998901111116", firstName: "Madina", lastName: "Nazarova", username: "madina_n"},
			{telegramID: 123456795, phone: "+998901111117", firstName: "Jasur", lastName: "Ortiqov", username: "jasur_o"},
			{telegramID: 123456796, phone: "+998901111118", firstName: "Nilufar", lastName: "Tursunova", username: "nilufar_t"},
			{telegramID: 123456797, phone: "+998901111119", firstName: "Sardor", lastName: "Abdullaev", username: "sardor_a"},
			{telegramID: 123456798, phone: "+998901111120", firstName: "Malika", lastName: "Karimova", username: "malika_k"},
		},
		branches: []seedBranch{
			{code: "main", name: "Demo Bar - Main", address: "Amir Temur St 42, Tashkent", lat: 41.2995, lng: 69.2401, isActive: true},
			{code: "downtown", name: "Demo Bar - Downtown", address: "Afrosiyob St 8, Tashkent", lat: 41.3057, lng: 69.2801, isActive: true},
			{code: "riverside", name: "Demo Bar - Riverside", address: "Kichik Halqa Yuli 14, Tashkent", lat: 41.2871, lng: 69.2684, isActive: true},
			{code: "chilonzor", name: "Demo Bar - Chilonzor", address: "Chilonzor 19 kvartal, Tashkent", lat: 41.2752, lng: 69.2038, isActive: true},
			{code: "samarkand-darvoza", name: "Demo Bar - Samarkand Darvoza", address: "Qatortol St 2, Tashkent", lat: 41.3164, lng: 69.2128, isActive: true},
			{code: "airport-road", name: "Demo Bar - Airport Road", address: "Kushbegi St 118, Tashkent", lat: 41.2578, lng: 69.2816, isActive: true},
		},
		menu: buildDemoBarMenu(),
		staff: []seedStaff{
			{staffCode: "admin", name: "Bar Director", role: "director", isActive: true},
			{branchCode: "main", staffCode: "manager-main", name: "Main Branch Manager", role: "manager", isActive: true},
			{branchCode: "downtown", staffCode: "manager-downtown", name: "Downtown Branch Manager", role: "manager", isActive: true},
			{branchCode: "riverside", staffCode: "manager-riverside", name: "Riverside Branch Manager", role: "manager", isActive: true},
			{branchCode: "chilonzor", staffCode: "manager-chilonzor", name: "Chilonzor Branch Manager", role: "manager", isActive: true},
			{branchCode: "samarkand-darvoza", staffCode: "manager-samarkand", name: "Samarkand Darvoza Manager", role: "manager", isActive: true},
			{branchCode: "airport-road", staffCode: "manager-airport", name: "Airport Road Manager", role: "manager", isActive: true},
			{branchCode: "main", staffCode: "barista-main-1", name: "Main Barista 1", role: "barista", isActive: true},
			{branchCode: "main", staffCode: "barista-main-2", name: "Main Barista 2", role: "barista", isActive: true},
			{branchCode: "main", staffCode: "barista-main-3", name: "Main Barista 3", role: "barista", isActive: true},
			{branchCode: "main", staffCode: "barista-main-4", name: "Main Barista 4", role: "barista", isActive: true},
			{branchCode: "downtown", staffCode: "barista-downtown-1", name: "Downtown Barista 1", role: "barista", isActive: true},
			{branchCode: "downtown", staffCode: "barista-downtown-2", name: "Downtown Barista 2", role: "barista", isActive: true},
			{branchCode: "downtown", staffCode: "barista-downtown-3", name: "Downtown Barista 3", role: "barista", isActive: true},
			{branchCode: "downtown", staffCode: "barista-downtown-4", name: "Downtown Barista 4", role: "barista", isActive: true},
			{branchCode: "riverside", staffCode: "barista-riverside-1", name: "Riverside Barista 1", role: "barista", isActive: true},
			{branchCode: "riverside", staffCode: "barista-riverside-2", name: "Riverside Barista 2", role: "barista", isActive: true},
			{branchCode: "riverside", staffCode: "barista-riverside-3", name: "Riverside Barista 3", role: "barista", isActive: true},
			{branchCode: "chilonzor", staffCode: "barista-chilonzor-1", name: "Chilonzor Barista 1", role: "barista", isActive: true},
			{branchCode: "chilonzor", staffCode: "barista-chilonzor-2", name: "Chilonzor Barista 2", role: "barista", isActive: true},
			{branchCode: "chilonzor", staffCode: "barista-chilonzor-3", name: "Chilonzor Barista 3", role: "barista", isActive: true},
			{branchCode: "samarkand-darvoza", staffCode: "barista-samarkand-1", name: "Samarkand Barista 1", role: "barista", isActive: true},
			{branchCode: "samarkand-darvoza", staffCode: "barista-samarkand-2", name: "Samarkand Barista 2", role: "barista", isActive: true},
			{branchCode: "samarkand-darvoza", staffCode: "barista-samarkand-3", name: "Samarkand Barista 3", role: "barista", isActive: true},
			{branchCode: "airport-road", staffCode: "barista-airport-1", name: "Airport Barista 1", role: "barista", isActive: true},
			{branchCode: "airport-road", staffCode: "barista-airport-2", name: "Airport Barista 2", role: "barista", isActive: true},
			{branchCode: "airport-road", staffCode: "barista-airport-3", name: "Airport Barista 3", role: "barista", isActive: true},
		},
		orderPlan: seedOrderPlan{totalOrders: 120},
	}
}

func buildUrbanCoffeeStore() seedStore {
	return seedStore{
		code:        "urbancoffee",
		slug:        "urban-coffee",
		name:        "Urban Coffee",
		category:    "cafe",
		description: "Specialty coffee and breakfast",
		address:     "Buyuk Ipak Yuli 1, Tashkent",
		phone:       "+998901234568",
		logoURL:     "",
		users: []seedUser{
			{telegramID: 123456899, phone: "+998902111111", firstName: "Lola", lastName: "Saidova", username: "lola_s"},
			{telegramID: 123456900, phone: "+998902111112", firstName: "Bekzod", lastName: "Nurmuhamedov", username: "bekzod_n"},
			{telegramID: 123456901, phone: "+998902111113", firstName: "Umida", lastName: "Rashidova", username: "umida_r"},
			{telegramID: 123456902, phone: "+998902111114", firstName: "Mirjalol", lastName: "Hamidov", username: "mirjalol_h"},
			{telegramID: 123456903, phone: "+998902111115", firstName: "Shahlo", lastName: "Ibragimova", username: "shahlo_i"},
			{telegramID: 123456904, phone: "+998902111116", firstName: "Iskandar", lastName: "Raxmatov", username: "iskandar_r"},
		},
		branches: []seedBranch{
			{code: "main", name: "Urban Coffee - Main", address: "Buyuk Ipak Yuli 1, Tashkent", lat: 41.3142, lng: 69.2879, isActive: true},
			{code: "yunusabad", name: "Urban Coffee - Yunusabad", address: "Yunusabad 17, Tashkent", lat: 41.3647, lng: 69.2965, isActive: true},
			{code: "parkent", name: "Urban Coffee - Parkent", address: "Parkent St 12, Tashkent", lat: 41.2898, lng: 69.3212, isActive: true},
		},
		menu: buildUrbanCoffeeMenu(),
		staff: []seedStaff{
			{staffCode: "admin", name: "Coffee Director", role: "director", isActive: true},
			{branchCode: "main", staffCode: "manager-main", name: "Main Branch Manager", role: "manager", isActive: true},
			{branchCode: "yunusabad", staffCode: "manager-yunusabad", name: "Yunusabad Branch Manager", role: "manager", isActive: true},
			{branchCode: "parkent", staffCode: "manager-parkent", name: "Parkent Branch Manager", role: "manager", isActive: true},
			{branchCode: "main", staffCode: "barista-main-1", name: "Main Barista 1", role: "barista", isActive: true},
			{branchCode: "main", staffCode: "barista-main-2", name: "Main Barista 2", role: "barista", isActive: true},
			{branchCode: "main", staffCode: "barista-main-3", name: "Main Barista 3", role: "barista", isActive: true},
			{branchCode: "yunusabad", staffCode: "barista-yunusabad-1", name: "Yunusabad Barista 1", role: "barista", isActive: true},
			{branchCode: "yunusabad", staffCode: "barista-yunusabad-2", name: "Yunusabad Barista 2", role: "barista", isActive: true},
			{branchCode: "yunusabad", staffCode: "barista-yunusabad-3", name: "Yunusabad Barista 3", role: "barista", isActive: true},
			{branchCode: "parkent", staffCode: "barista-parkent-1", name: "Parkent Barista 1", role: "barista", isActive: true},
			{branchCode: "parkent", staffCode: "barista-parkent-2", name: "Parkent Barista 2", role: "barista", isActive: true},
			{branchCode: "parkent", staffCode: "barista-parkent-3", name: "Parkent Barista 3", role: "barista", isActive: false},
		},
		orderPlan: seedOrderPlan{totalOrders: 45},
	}
}

func buildStreetBurgerStore() seedStore {
	return seedStore{
		code:        "streetburger",
		slug:        "street-burger",
		name:        "Street Burger",
		category:    "fastfood",
		description: "Burgers, chicken, and quick comfort food",
		address:     "Bodomzor 3, Tashkent",
		phone:       "+998901234569",
		logoURL:     "",
		users: []seedUser{
			{telegramID: 123456999, phone: "+998903111111", firstName: "Akmal", lastName: "Ganiev", username: "akmal_g"},
			{telegramID: 123457000, phone: "+998903111112", firstName: "Sevara", lastName: "Mirzaeva", username: "sevara_m"},
			{telegramID: 123457001, phone: "+998903111113", firstName: "Timur", lastName: "Kholmatov", username: "timur_k"},
			{telegramID: 123457002, phone: "+998903111114", firstName: "Sabina", lastName: "Khamidova", username: "sabina_k"},
			{telegramID: 123457003, phone: "+998903111115", firstName: "Shavkat", lastName: "Akhmedov", username: "shavkat_a"},
			{telegramID: 123457004, phone: "+998903111116", firstName: "Gulnora", lastName: "Mamatova", username: "gulnora_m"},
		},
		branches: []seedBranch{
			{code: "main", name: "Street Burger - Main", address: "Bodomzor 3, Tashkent", lat: 41.3471, lng: 69.2872, isActive: true},
			{code: "almazar", name: "Street Burger - Almazar", address: "Almazar 8, Tashkent", lat: 41.3502, lng: 69.2198, isActive: true},
			{code: "sergeli", name: "Street Burger - Sergeli", address: "Sergeli 9, Tashkent", lat: 41.2259, lng: 69.2144, isActive: true},
		},
		menu: buildStreetBurgerMenu(),
		staff: []seedStaff{
			{staffCode: "admin", name: "Burger Director", role: "director", isActive: true},
			{branchCode: "main", staffCode: "manager-main", name: "Main Branch Manager", role: "manager", isActive: true},
			{branchCode: "almazar", staffCode: "manager-almazar", name: "Almazar Branch Manager", role: "manager", isActive: true},
			{branchCode: "sergeli", staffCode: "manager-sergeli", name: "Sergeli Branch Manager", role: "manager", isActive: true},
			{branchCode: "main", staffCode: "barista-main-1", name: "Main Barista 1", role: "barista", isActive: true},
			{branchCode: "main", staffCode: "barista-main-2", name: "Main Barista 2", role: "barista", isActive: true},
			{branchCode: "main", staffCode: "barista-main-3", name: "Main Barista 3", role: "barista", isActive: true},
			{branchCode: "almazar", staffCode: "barista-almazar-1", name: "Almazar Barista 1", role: "barista", isActive: true},
			{branchCode: "almazar", staffCode: "barista-almazar-2", name: "Almazar Barista 2", role: "barista", isActive: true},
			{branchCode: "almazar", staffCode: "barista-almazar-3", name: "Almazar Barista 3", role: "barista", isActive: true},
			{branchCode: "sergeli", staffCode: "barista-sergeli-1", name: "Sergeli Barista 1", role: "barista", isActive: true},
			{branchCode: "sergeli", staffCode: "barista-sergeli-2", name: "Sergeli Barista 2", role: "barista", isActive: true},
			{branchCode: "sergeli", staffCode: "barista-sergeli-3", name: "Sergeli Barista 3", role: "barista", isActive: true},
		},
		orderPlan: seedOrderPlan{totalOrders: 45},
	}
}

func buildDemoBarMenu() seedMenu {
	return seedMenu{
		categories: []seedCategory{
			{
				name: "Cocktails",
				items: []seedCatalogItem{
					{
						name:        "Mojito",
						description: "Fresh lime, mint, rum, soda",
						price:       45000,
						modGroups: []seedModifierGroup{
							{
								name:          "Size",
								selectionType: "single",
								required:      true,
								mods: []seedModifier{
									{name: "Regular", price: 0},
									{name: "Large", price: 15000},
								},
							},
							{
								name:          "Extras",
								selectionType: "multiple",
								required:      false,
								mods: []seedModifier{
									{name: "Extra mint", price: 3000},
									{name: "Extra lime", price: 3000},
									{name: "Double shot", price: 10000},
								},
							},
						},
						unavailableAt: map[string]bool{
							"airport-road": true,
						},
					},
					{
						name:        "Margarita",
						description: "Tequila, lime juice, triple sec",
						price:       50000,
						modGroups: []seedModifierGroup{
							{
								name:          "Size",
								selectionType: "single",
								required:      true,
								mods: []seedModifier{
									{name: "Regular", price: 0},
									{name: "Large", price: 15000},
								},
							},
							{
								name:          "Salt rim",
								selectionType: "single",
								required:      false,
								mods: []seedModifier{
									{name: "With salt", price: 0},
									{name: "No salt", price: 0},
								},
							},
						},
					},
					{
						name:        "Long Island",
						description: "Five spirits, cola, lemon",
						price:       55000,
						modGroups: []seedModifierGroup{
							{
								name:          "Size",
								selectionType: "single",
								required:      true,
								mods: []seedModifier{
									{name: "Regular", price: 0},
									{name: "Large", price: 20000},
								},
							},
						},
					},
					{
						name:        "Aperol Spritz",
						description: "Aperol, prosecco, soda",
						price:       48000,
					},
					{
						name:        "Negroni",
						description: "Gin, vermouth, campari",
						price:       52000,
					},
				},
			},
			{
				name: "Beer",
				items: []seedCatalogItem{
					{
						name:        "Sarbast Lager",
						description: "Local Uzbek lager, crisp and light",
						price:       25000,
						modGroups: []seedModifierGroup{
							{
								name:          "Size",
								selectionType: "single",
								required:      true,
								mods: []seedModifier{
									{name: "0.33L", price: 0},
									{name: "0.5L", price: 10000},
								},
							},
						},
					},
					{
						name:        "Pulsar IPA",
						description: "Hoppy craft IPA",
						price:       35000,
						modGroups: []seedModifierGroup{
							{
								name:          "Size",
								selectionType: "single",
								required:      true,
								mods: []seedModifier{
									{name: "0.33L", price: 0},
									{name: "0.5L", price: 12000},
								},
							},
						},
					},
					{
						name:        "Corona",
						description: "Mexican lager with lime",
						price:       40000,
					},
					{
						name:        "Hoegaarden",
						description: "Wheat beer with citrus notes",
						price:       42000,
					},
				},
			},
			{
				name: "Wine",
				items: []seedCatalogItem{
					{name: "House Red", description: "Full-bodied red blend", price: 45000},
					{name: "House White", description: "Crisp white wine", price: 45000},
					{name: "Sparkling Brut", description: "Dry sparkling wine", price: 58000},
					{name: "Rosé", description: "Light and fresh rosé", price: 50000},
				},
			},
			{
				name: "Coffee",
				items: []seedCatalogItem{
					{name: "Espresso", description: "Single shot espresso", price: 12000},
					{name: "Americano", description: "Espresso with hot water", price: 14000},
					{name: "Cappuccino", description: "Espresso, steamed milk, foam", price: 18000},
					{name: "Latte", description: "Espresso, milk, light foam", price: 20000},
					{name: "Flat White", description: "Velvety milk coffee", price: 19000},
				},
			},
			{
				name: "Snacks",
				items: []seedCatalogItem{
					{
						name:        "Nachos",
						description: "Tortilla chips with cheese sauce and jalapenos",
						price:       35000,
						modGroups: []seedModifierGroup{
							{
								name:          "Toppings",
								selectionType: "multiple",
								required:      false,
								mods: []seedModifier{
									{name: "Extra cheese", price: 5000},
									{name: "Guacamole", price: 8000},
									{name: "Sour cream", price: 5000},
								},
							},
						},
					},
					{
						name:        "Chicken Wings",
						description: "Crispy wings with sauce",
						price:       42000,
						modGroups: []seedModifierGroup{
							{
								name:          "Sauce",
								selectionType: "single",
								required:      true,
								mods: []seedModifier{
									{name: "BBQ", price: 0},
									{name: "Buffalo", price: 0},
									{name: "Honey Mustard", price: 0},
								},
							},
							{
								name:          "Portion",
								selectionType: "single",
								required:      true,
								mods: []seedModifier{
									{name: "6 pieces", price: 0},
									{name: "12 pieces", price: 25000},
								},
							},
						},
					},
					{
						name:        "French Fries",
						description: "Crispy golden fries",
						price:       20000,
						modGroups: []seedModifierGroup{
							{
								name:          "Dip",
								selectionType: "single",
								required:      false,
								mods: []seedModifier{
									{name: "Ketchup", price: 0},
									{name: "Mayo", price: 0},
									{name: "Cheese sauce", price: 3000},
								},
							},
						},
					},
					{name: "Olives", description: "Marinated green and black olives", price: 18000},
					{name: "Truffle Fries", description: "Fries with truffle salt", price: 24000},
				},
			},
			{
				name: "Sharing Plates",
				items: []seedCatalogItem{
					{name: "Cheese Board", description: "Mixed cheeses, crackers, fruit", price: 68000},
					{name: "Mezze Platter", description: "Hummus, pita, dips, vegetables", price: 62000},
					{name: "Bruschetta", description: "Tomato, basil, olive oil on toasted bread", price: 28000},
					{name: "Mini Sliders", description: "Three small beef sliders", price: 54000},
					{name: "Fried Calamari", description: "Lightly fried calamari rings", price: 58000},
				},
			},
		},
	}
}

func buildUrbanCoffeeMenu() seedMenu {
	return seedMenu{
		categories: []seedCategory{
			{
				name: "Espresso Bar",
				items: []seedCatalogItem{
					{name: "Espresso", description: "Strong single shot", price: 12000},
					{name: "Double Espresso", description: "Double shot", price: 16000},
					{name: "Americano", description: "Espresso and hot water", price: 14000},
					{name: "Flat White", description: "Velvety milk coffee", price: 18000},
					{name: "Cappuccino", description: "Espresso with milk foam", price: 19000},
				},
			},
			{
				name: "Signature Drinks",
				items: []seedCatalogItem{
					{name: "Iced Latte", description: "Cold latte over ice", price: 22000},
					{name: "Caramel Macchiato", description: "Espresso with caramel syrup", price: 24000},
					{name: "Vanilla Raf", description: "Creamy vanilla coffee", price: 26000},
					{name: "Mocha", description: "Coffee with chocolate", price: 23000},
					{name: "Cold Brew", description: "Slow-steeped iced coffee", price: 25000},
				},
			},
			{
				name: "Tea",
				items: []seedCatalogItem{
					{name: "Black Tea", description: "Classic black tea", price: 10000},
					{name: "Green Tea", description: "Light green tea", price: 10000},
					{name: "Jasmine Tea", description: "Floral jasmine tea", price: 12000},
					{name: "Berry Tea", description: "Berry blend tea", price: 14000},
				},
			},
			{
				name: "Pastries",
				items: []seedCatalogItem{
					{name: "Croissant", description: "Buttery croissant", price: 14000},
					{name: "Chocolate Muffin", description: "Rich chocolate muffin", price: 16000},
					{name: "Cinnamon Roll", description: "Warm cinnamon roll", price: 18000},
					{name: "Banana Bread", description: "Moist banana loaf slice", price: 17000},
					{name: "Cheesecake Slice", description: "Classic cheesecake", price: 24000},
				},
			},
			{
				name: "Breakfast",
				items: []seedCatalogItem{
					{name: "Avocado Toast", description: "Avocado on sourdough", price: 28000},
					{name: "Egg Sandwich", description: "Egg and cheese sandwich", price: 26000},
					{name: "Oatmeal Bowl", description: "Oats with fruit and honey", price: 22000},
					{name: "Granola Yogurt", description: "Yogurt with granola", price: 24000},
				},
			},
		},
	}
}

func buildStreetBurgerMenu() seedMenu {
	return seedMenu{
		categories: []seedCategory{
			{
				name: "Burgers",
				items: []seedCatalogItem{
					{
						name:        "Classic Burger",
						description: "Beef patty, lettuce, tomato, pickles",
						price:       45000,
						modGroups: []seedModifierGroup{
							{
								name:          "Doneness",
								selectionType: "single",
								required:      false,
								mods: []seedModifier{
									{name: "Medium", price: 0},
									{name: "Well Done", price: 0},
								},
							},
						},
					},
					{name: "Cheeseburger", description: "Burger with cheddar", price: 48000},
					{name: "Double Burger", description: "Two patties and cheese", price: 55000},
					{name: "BBQ Burger", description: "BBQ sauce and onion rings", price: 50000},
					{name: "Mushroom Burger", description: "Mushrooms and creamy sauce", price: 52000},
				},
			},
			{
				name: "Chicken",
				items: []seedCatalogItem{
					{name: "Chicken Burger", description: "Crispy chicken burger", price: 42000},
					{name: "Spicy Chicken Burger", description: "Spicy chicken and slaw", price: 44000},
					{name: "Chicken Tenders", description: "Breaded chicken strips", price: 38000},
					{name: "Chicken Wings", description: "Sauced wings", price: 40000},
				},
			},
			{
				name: "Sides",
				items: []seedCatalogItem{
					{name: "Fries", description: "Crispy fries", price: 16000},
					{name: "Onion Rings", description: "Golden onion rings", price: 18000},
					{name: "Coleslaw", description: "Creamy cabbage slaw", price: 12000},
					{name: "Nuggets", description: "Chicken nuggets", price: 22000},
				},
			},
			{
				name: "Combos",
				items: []seedCatalogItem{
					{name: "Burger Combo", description: "Burger, fries, drink", price: 62000},
					{name: "Chicken Combo", description: "Chicken meal with fries", price: 60000},
					{name: "Family Box", description: "Mixed burgers and sides", price: 145000, unavailableAt: map[string]bool{"sergeli": true}},
				},
			},
			{
				name: "Drinks",
				items: []seedCatalogItem{
					{name: "Cola", description: "Classic cola", price: 12000},
					{name: "Lemonade", description: "Fresh lemonade", price: 14000},
					{name: "Iced Tea", description: "Cold iced tea", price: 13000},
					{name: "Mineral Water", description: "Still mineral water", price: 8000},
				},
			},
		},
	}
}

func seedBranchMenu(ctx context.Context, conn *pgx.Conn, storeID, branchID string, store seedStore, branch seedBranch) error {
	for categoryIndex, category := range store.menu.categories {
		categoryID, err := ensureCategory(ctx, conn, storeID, branchID, category.name, categoryIndex)
		if err != nil {
			return err
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
				return err
			}

			if err := setItemAvailability(ctx, conn, itemID, !catalogItem.unavailableAt[branch.code]); err != nil {
				return err
			}

			for groupIndex, group := range catalogItem.modGroups {
				groupID, err := ensureModifierGroup(ctx, conn, storeID, branchID, itemID, group, groupIndex)
				if err != nil {
					return err
				}

				for modIndex, mod := range group.mods {
					if err := ensureModifier(ctx, conn, storeID, branchID, groupID, mod, modIndex); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func setItemAvailability(ctx context.Context, conn *pgx.Conn, itemID string, isAvailable bool) error {
	_, err := conn.Exec(ctx, `
		UPDATE items
		SET is_available = $2
		WHERE id = $1
	`, itemID, isAvailable)
	return err
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, true)
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
	return id, err
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

func ensureModifier(ctx context.Context, conn *pgx.Conn, storeID, branchID, groupID string, mod seedModifier, sortOrder int) error {
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
		return err
	}
	if err != pgx.ErrNoRows {
		return err
	}

	_, err = conn.Exec(ctx, `
		INSERT INTO modifiers (modifier_group_id, store_id, branch_id, name, price_adjustment, is_available, sort_order)
		VALUES ($1, $2, $3, $4, $5, true, $6)
	`, groupID, storeID, branchID, mod.name, mod.price, sortOrder)
	return err
}

func replaceSeedOrders(ctx context.Context, conn *pgx.Conn, store seedStore, refs seededStoreRefs, userIDs []string) error {
	if len(userIDs) == 0 {
		return fmt.Errorf("no seeded users configured for store %s", store.code)
	}

	_, err := conn.Exec(ctx, `
		DELETE FROM orders
		WHERE store_id = $1
		  AND user_id = ANY($2::uuid[])
	`, refs.storeID, userIDs)
	if err != nil {
		return err
	}

	specs := generateSeedOrders(store, refs, userIDs)
	for _, spec := range specs {
		if err := insertSeedOrder(ctx, conn, spec); err != nil {
			return err
		}
	}
	return nil
}

func generateSeedOrders(store seedStore, refs seededStoreRefs, userIDs []string) []seedGeneratedOrder {
	branchCodes := make([]string, 0, len(refs.branchIDs))
	for code := range refs.branchIDs {
		branchCodes = append(branchCodes, code)
	}
	sort.Strings(branchCodes)

	statuses := buildStatusPlan(store.code, store.orderPlan.totalOrders)

	orders := make([]seedGeneratedOrder, 0, store.orderPlan.totalOrders)
	baseTime := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)
	for i := 0; i < store.orderPlan.totalOrders; i++ {
		status := "picked_up"
		if i < len(statuses) {
			status = statuses[i]
		}
		branchCode := branchCodes[i%len(branchCodes)]
		branchID := refs.branchIDs[branchCode]
		userID := userIDs[i%len(userIDs)]
		total := int64(42000 + (i%7)*5000)
		orders = append(orders, seedGeneratedOrder{
			userID:          userID,
			storeID:         refs.storeID,
			branchID:        branchID,
			status:          status,
			paymentMethod:   "pay_at_pickup",
			paymentStatus:   paymentStatusFor(status),
			etaMinutes:      10 + (i % 8),
			totalPrice:      total,
			rejectionReason: rejectionReasonFor(status),
			createdAt:       baseTime.Add(-time.Duration(i) * 3 * time.Hour),
			items: []seedGeneratedOrderItem{
				{
					itemName:  fmt.Sprintf("%s Special %d", store.name, i+1),
					itemPrice: total,
					quantity:  1 + (i % 3),
				},
			},
		})
	}
	return orders
}

func buildStatusPlan(storeCode string, total int) []string {
	type statusBucket struct {
		status string
		count  int
	}

	var buckets []statusBucket
	switch storeCode {
	case "demobar":
		buckets = []statusBucket{
			{status: "pending", count: 4},
			{status: "accepted", count: 5},
			{status: "preparing", count: 6},
			{status: "ready", count: 5},
			{status: "cancelled", count: 4},
			{status: "rejected", count: 2},
		}
	default:
		buckets = []statusBucket{
			{status: "pending", count: 2},
			{status: "accepted", count: 2},
			{status: "preparing", count: 2},
			{status: "ready", count: 2},
			{status: "cancelled", count: 2},
			{status: "rejected", count: 1},
		}
	}

	statuses := make([]string, 0, total)
	for _, bucket := range buckets {
		statuses = append(statuses, repeatStatus(bucket.status, bucket.count)...)
	}
	if len(statuses) < total {
		statuses = append(statuses, repeatStatus("picked_up", total-len(statuses))...)
	}
	if len(statuses) > total {
		statuses = statuses[:total]
	}
	return statuses
}

func insertSeedOrder(ctx context.Context, conn *pgx.Conn, order seedGeneratedOrder) error {
	var orderID string
	err := conn.QueryRow(ctx, `
		INSERT INTO orders (
			user_id, store_id, branch_id, status, total_price, payment_method, payment_status,
			eta_minutes, rejection_reason, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
		RETURNING id
	`, order.userID, order.storeID, order.branchID, order.status, order.totalPrice, order.paymentMethod, order.paymentStatus, order.etaMinutes, order.rejectionReason, order.createdAt).Scan(&orderID)
	if err != nil {
		return err
	}

	for _, item := range order.items {
		itemID, err := insertSeedOrderItem(ctx, conn, orderID, item)
		if err != nil {
			return err
		}
		for _, mod := range item.modifiers {
			if err := insertSeedOrderModifier(ctx, conn, itemID, mod); err != nil {
				return err
			}
		}
	}
	return nil
}

func insertSeedOrderItem(ctx context.Context, conn *pgx.Conn, orderID string, item seedGeneratedOrderItem) (string, error) {
	var id string
	err := conn.QueryRow(ctx, `
		INSERT INTO order_items (order_id, item_id, item_name, item_price, quantity)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, orderID, item.itemID, item.itemName, item.itemPrice, item.quantity).Scan(&id)
	return id, err
}

func insertSeedOrderModifier(ctx context.Context, conn *pgx.Conn, orderItemID string, mod seedGeneratedOrderModifier) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO order_item_modifiers (order_item_id, modifier_id, modifier_name, price_adjustment)
		VALUES ($1, $2, $3, $4)
	`, orderItemID, mod.modifierID, mod.modifierName, mod.priceAdjustment)
	return err
}

func repeatStatus(status string, count int) []string {
	result := make([]string, 0, count)
	for i := 0; i < count; i++ {
		result = append(result, status)
	}
	return result
}

func paymentStatusFor(status string) string {
	switch status {
	case "picked_up":
		return "paid"
	case "cancelled", "rejected":
		return "failed"
	default:
		return "pending"
	}
}

func rejectionReasonFor(status string) string {
	if status == "rejected" {
		return "Kitchen capacity reached for the seeded demo window"
	}
	return ""
}
