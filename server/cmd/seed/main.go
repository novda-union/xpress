package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/xpressgo/server/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()

	conn, err := pgx.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	ctx := context.Background()

	// Hash password for demo staff
	hash, err := bcrypt.GenerateFromPassword([]byte("demo123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	// Create demo store
	var storeID string
	err = conn.QueryRow(ctx, `
		INSERT INTO stores (name, code, slug, description, address, phone, commission_rate)
		VALUES ('Demo Bar', 'demobar', 'demo-bar', 'The best cocktails in Tashkent', 'Amir Temur St 42, Tashkent', '+998901234567', 5.00)
		ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`).Scan(&storeID)
	if err != nil {
		log.Fatalf("failed to create store: %v", err)
	}
	log.Printf("Store created: %s", storeID)

	// Create staff
	_, err = conn.Exec(ctx, `
		INSERT INTO store_staff (store_id, staff_code, name, password_hash, role)
		VALUES ($1, 'admin', 'Bar Manager', $2, 'owner')
		ON CONFLICT (store_id, staff_code) DO NOTHING
	`, storeID, string(hash))
	if err != nil {
		log.Fatalf("failed to create staff: %v", err)
	}
	log.Println("Staff created: admin")

	// Create demo user
	_, err = conn.Exec(ctx, `
		INSERT INTO users (telegram_id, phone, first_name, last_name, username)
		VALUES (123456789, '+998901111111', 'Demo', 'User', 'demouser')
		ON CONFLICT (telegram_id) DO NOTHING
	`)
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}
	log.Println("Demo user created")

	// Categories
	type catInfo struct {
		name string
		id   string
	}
	categories := []string{"Cocktails", "Beer", "Snacks"}
	cats := make([]catInfo, 0, len(categories))

	for i, name := range categories {
		var catID string
		err = conn.QueryRow(ctx, `
			INSERT INTO categories (store_id, name, sort_order)
			VALUES ($1, $2, $3)
			RETURNING id
		`, storeID, name, i).Scan(&catID)
		if err != nil {
			log.Fatalf("failed to create category %s: %v", name, err)
		}
		cats = append(cats, catInfo{name: name, id: catID})
	}
	log.Printf("Created %d categories", len(cats))

	// Items with modifiers
	items := []struct {
		category    int
		name        string
		description string
		price       int64
		modGroups   []struct {
			name          string
			selectionType string
			required      bool
			mods          []struct {
				name  string
				price int64
			}
		}
	}{
		// Cocktails
		{0, "Mojito", "Fresh lime, mint, rum, soda", 45000, []struct {
			name          string
			selectionType string
			required      bool
			mods          []struct {
				name  string
				price int64
			}
		}{
			{"Size", "single", true, []struct {
				name  string
				price int64
			}{{"Regular", 0}, {"Large", 15000}}},
			{"Extras", "multiple", false, []struct {
				name  string
				price int64
			}{{"Extra mint", 3000}, {"Extra lime", 3000}, {"Double shot", 10000}}},
		}},
		{0, "Margarita", "Tequila, lime juice, triple sec", 50000, []struct {
			name          string
			selectionType string
			required      bool
			mods          []struct {
				name  string
				price int64
			}
		}{
			{"Size", "single", true, []struct {
				name  string
				price int64
			}{{"Regular", 0}, {"Large", 15000}}},
			{"Salt rim", "single", false, []struct {
				name  string
				price int64
			}{{"With salt", 0}, {"No salt", 0}}},
		}},
		{0, "Long Island", "Five spirits, cola, lemon", 55000, []struct {
			name          string
			selectionType string
			required      bool
			mods          []struct {
				name  string
				price int64
			}
		}{
			{"Size", "single", true, []struct {
				name  string
				price int64
			}{{"Regular", 0}, {"Large", 20000}}},
		}},
		{0, "Aperol Spritz", "Aperol, prosecco, soda", 48000, nil},

		// Beer
		{1, "Sarbast Lager", "Local Uzbek lager, crisp and light", 25000, []struct {
			name          string
			selectionType string
			required      bool
			mods          []struct {
				name  string
				price int64
			}
		}{
			{"Size", "single", true, []struct {
				name  string
				price int64
			}{{"0.33L", 0}, {"0.5L", 10000}}},
		}},
		{1, "Pulsar IPA", "Hoppy craft IPA", 35000, []struct {
			name          string
			selectionType string
			required      bool
			mods          []struct {
				name  string
				price int64
			}
		}{
			{"Size", "single", true, []struct {
				name  string
				price int64
			}{{"0.33L", 0}, {"0.5L", 12000}}},
		}},
		{1, "Corona", "Mexican lager with lime", 40000, nil},

		// Snacks
		{2, "Nachos", "Tortilla chips with cheese sauce and jalapenos", 35000, []struct {
			name          string
			selectionType string
			required      bool
			mods          []struct {
				name  string
				price int64
			}
		}{
			{"Toppings", "multiple", false, []struct {
				name  string
				price int64
			}{{"Extra cheese", 5000}, {"Guacamole", 8000}, {"Sour cream", 5000}}},
		}},
		{2, "Chicken Wings", "Crispy wings with sauce", 42000, []struct {
			name          string
			selectionType string
			required      bool
			mods          []struct {
				name  string
				price int64
			}
		}{
			{"Sauce", "single", true, []struct {
				name  string
				price int64
			}{{"BBQ", 0}, {"Buffalo", 0}, {"Honey Mustard", 0}}},
			{"Portion", "single", true, []struct {
				name  string
				price int64
			}{{"6 pieces", 0}, {"12 pieces", 25000}}},
		}},
		{2, "French Fries", "Crispy golden fries", 20000, []struct {
			name          string
			selectionType string
			required      bool
			mods          []struct {
				name  string
				price int64
			}
		}{
			{"Dip", "single", false, []struct {
				name  string
				price int64
			}{{"Ketchup", 0}, {"Mayo", 0}, {"Cheese sauce", 3000}}},
		}},
	}

	for i, item := range items {
		var itemID string
		err = conn.QueryRow(ctx, `
			INSERT INTO items (category_id, store_id, name, description, base_price, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, cats[item.category].id, storeID, item.name, item.description, item.price, i).Scan(&itemID)
		if err != nil {
			log.Fatalf("failed to create item %s: %v", item.name, err)
		}

		for gi, mg := range item.modGroups {
			var groupID string
			minSel := 0
			maxSel := 1
			if mg.required {
				minSel = 1
			}
			if mg.selectionType == "multiple" {
				maxSel = len(mg.mods)
			}
			err = conn.QueryRow(ctx, `
				INSERT INTO modifier_groups (item_id, store_id, name, selection_type, is_required, min_selections, max_selections, sort_order)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id
			`, itemID, storeID, mg.name, mg.selectionType, mg.required, minSel, maxSel, gi).Scan(&groupID)
			if err != nil {
				log.Fatalf("failed to create modifier group %s: %v", mg.name, err)
			}

			for mi, mod := range mg.mods {
				_, err = conn.Exec(ctx, `
					INSERT INTO modifiers (modifier_group_id, store_id, name, price_adjustment, sort_order)
					VALUES ($1, $2, $3, $4, $5)
				`, groupID, storeID, mod.name, mod.price, mi)
				if err != nil {
					log.Fatalf("failed to create modifier %s: %v", mod.name, err)
				}
			}
		}
	}
	log.Printf("Created %d items with modifiers", len(items))

	log.Println("Seed completed successfully!")
}
