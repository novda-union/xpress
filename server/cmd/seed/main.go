package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/xpressgo/server/internal/config"
	"github.com/xpressgo/server/internal/database"
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
	totalOrders int
}

func main() {
	cfg := config.Load()

	conn, err := database.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	ctx := context.Background()

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

		for _, branch := range store.branches {
			branchID, err := ensureBranch(ctx, conn, storeID, branch)
			if err != nil {
				log.Fatalf("failed to ensure branch %s/%s: %v", store.code, branch.code, err)
			}
			log.Printf("Branch ready: %s (%s)", branch.name, branchID)
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
		branches: []seedBranch{
			{code: "main", name: "Demo Bar - Main", address: "Amir Temur St 42, Tashkent", lat: 41.2995, lng: 69.2401, isActive: true},
			{code: "downtown", name: "Demo Bar - Downtown", address: "Afrosiyob St 8, Tashkent", lat: 41.3057, lng: 69.2801, isActive: true},
			{code: "riverside", name: "Demo Bar - Riverside", address: "Kichik Halqa Yuli 14, Tashkent", lat: 41.2871, lng: 69.2684, isActive: true},
			{code: "chilonzor", name: "Demo Bar - Chilonzor", address: "Chilonzor 19 kvartal, Tashkent", lat: 41.2752, lng: 69.2038, isActive: true},
			{code: "samarkand-darvoza", name: "Demo Bar - Samarkand Darvoza", address: "Qatortol St 2, Tashkent", lat: 41.3164, lng: 69.2128, isActive: true},
			{code: "airport-road", name: "Demo Bar - Airport Road", address: "Kushbegi St 118, Tashkent", lat: 41.2578, lng: 69.2816, isActive: true},
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
		branches: []seedBranch{
			{code: "main", name: "Urban Coffee - Main", address: "Buyuk Ipak Yuli 1, Tashkent", lat: 41.3142, lng: 69.2879, isActive: true},
			{code: "yunusabad", name: "Urban Coffee - Yunusabad", address: "Yunusabad 17, Tashkent", lat: 41.3647, lng: 69.2965, isActive: true},
			{code: "parkent", name: "Urban Coffee - Parkent", address: "Parkent St 12, Tashkent", lat: 41.2898, lng: 69.3212, isActive: true},
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
		branches: []seedBranch{
			{code: "main", name: "Street Burger - Main", address: "Bodomzor 3, Tashkent", lat: 41.3471, lng: 69.2872, isActive: true},
			{code: "almazar", name: "Street Burger - Almazar", address: "Almazar 8, Tashkent", lat: 41.3502, lng: 69.2198, isActive: true},
			{code: "sergeli", name: "Street Burger - Sergeli", address: "Sergeli 9, Tashkent", lat: 41.2259, lng: 69.2144, isActive: true},
		},
		orderPlan: seedOrderPlan{totalOrders: 45},
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

func ensureUser(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO users (telegram_id, phone, first_name, last_name, username)
		VALUES (123456789, '+998901111111', 'Demo', 'User', 'demouser')
		ON CONFLICT (telegram_id) DO UPDATE SET
			phone = EXCLUDED.phone,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			username = EXCLUDED.username
	`)
	return err
}

func ensureStaff(ctx context.Context, conn *pgx.Conn, storeID string, branchID *string, staffCode, name, passwordHash, role string) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO store_staff (store_id, branch_id, staff_code, name, password_hash, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (store_id, staff_code) DO UPDATE SET
			branch_id = EXCLUDED.branch_id,
			name = EXCLUDED.name,
			password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role,
			is_active = true
	`, storeID, branchID, staffCode, name, passwordHash, role)
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
