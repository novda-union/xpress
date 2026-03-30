package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/xpressgo/server/internal/config"
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

func main() {
	cfg := config.Load()

	conn, err := pgx.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte("demo123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	storeID, err := ensureStore(ctx, conn)
	if err != nil {
		log.Fatalf("failed to create store: %v", err)
	}
	log.Printf("Store ready: %s", storeID)

	branchID, err := ensureBranch(ctx, conn, storeID)
	if err != nil {
		log.Fatalf("failed to create branch: %v", err)
	}
	log.Printf("Branch ready: %s", branchID)

	if err := ensureStaff(ctx, conn, storeID, nil, "admin", "Bar Manager", string(hash), "director"); err != nil {
		log.Fatalf("failed to create staff: %v", err)
	}
	branchScopedID := branchID
	if err := ensureStaff(ctx, conn, storeID, &branchScopedID, "manager", "Branch Manager", string(hash), "manager"); err != nil {
		log.Fatalf("failed to create branch staff: %v", err)
	}
	log.Println("Staff ready: admin, manager")

	if err := ensureUser(ctx, conn); err != nil {
		log.Fatalf("failed to create user: %v", err)
	}
	log.Println("Demo user ready")

	categoryIDs := map[string]string{}
	categoryNames := []string{"Cocktails", "Beer", "Snacks"}
	for i, name := range categoryNames {
		id, err := ensureCategory(ctx, conn, storeID, branchID, name, i)
		if err != nil {
			log.Fatalf("failed to create category %s: %v", name, err)
		}
		categoryIDs[name] = id
	}
	log.Printf("Created %d categories", len(categoryIDs))

	items := []seedItem{
		{
			category:    "Cocktails",
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
		},
		{
			category:    "Cocktails",
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
			category:    "Cocktails",
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
			category:    "Cocktails",
			name:        "Aperol Spritz",
			description: "Aperol, prosecco, soda",
			price:       48000,
		},
		{
			category:    "Beer",
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
			category:    "Beer",
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
			category:    "Beer",
			name:        "Corona",
			description: "Mexican lager with lime",
			price:       40000,
		},
		{
			category:    "Snacks",
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
			category:    "Snacks",
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
			category:    "Snacks",
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
	}

	for i, item := range items {
		categoryID := categoryIDs[item.category]
		itemID, err := ensureItem(ctx, conn, storeID, branchID, categoryID, item, i)
		if err != nil {
			log.Fatalf("failed to create item %s: %v", item.name, err)
		}

		for gi, mg := range item.modGroups {
			groupID, err := ensureModifierGroup(ctx, conn, storeID, branchID, itemID, mg, gi)
			if err != nil {
				log.Fatalf("failed to create modifier group %s: %v", mg.name, err)
			}

			for mi, mod := range mg.mods {
				if err := ensureModifier(ctx, conn, storeID, branchID, groupID, mod, mi); err != nil {
					log.Fatalf("failed to create modifier %s: %v", mod.name, err)
				}
			}
		}
	}
	log.Printf("Created %d items with modifiers", len(items))

	log.Println("Seed completed successfully!")
}

func ensureStore(ctx context.Context, conn *pgx.Conn) (string, error) {
	var id string
	err := conn.QueryRow(ctx, `
		INSERT INTO stores (
			name, code, slug, category, description, address, phone, logo_url,
			telegram_group_chat_id, subscription_tier, commission_rate, is_active
		)
		VALUES (
			'Demo Bar', 'demobar', 'demo-bar', 'bar', 'The best cocktails in Tashkent',
			'Amir Temur St 42, Tashkent', '+998901234567', '',
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
	`).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func ensureBranch(ctx context.Context, conn *pgx.Conn, storeID string) (string, error) {
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
	`, storeID, "Demo Bar - Main", "Amir Temur St 42, Tashkent", 41.2995, 69.2401, "", nil).Scan(&id)
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
