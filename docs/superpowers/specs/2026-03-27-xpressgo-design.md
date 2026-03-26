# Xpressgo — Design Spec

## Overview

Xpressgo is a Telegram-based ordering platform for bars, restaurants, and coffee shops in Tashkent, Uzbekistan. Users order via a Telegram Mini App, pick up when ready — eliminating lines and wait times. B2B SaaS model: monthly subscriptions + transaction commissions.

**Prototype target:** Demo-ready in 1-2 weeks, running locally via Docker Compose. First client: a bar in Tashkent.

## Architecture

**Monorepo, single Go server.** One binary handles REST API, WebSocket hub, and Telegram bot.

```
xpressgo/
├── server/          # Go (Echo) — single binary
│   ├── cmd/server/          # main.go entry point
│   ├── internal/
│   │   ├── config/          # env config loading
│   │   ├── handler/         # Echo HTTP handlers
│   │   ├── middleware/       # auth, tenant context, CORS
│   │   ├── model/           # Go structs matching DB tables
│   │   ├── repository/      # DB queries (one file per entity)
│   │   ├── service/         # Business logic
│   │   ├── ws/              # WebSocket hub
│   │   └── telegram/        # Bot — auth, notifications, store alerts
│   ├── migrations/          # SQL migration files
│   └── pkg/                 # Shared utilities
├── web/             # React (Telegram Mini App)
│   ├── src/
│   │   ├── app/              # Router, providers
│   │   ├── pages/            # StorePage, ItemPage, CartPage, OrderPage, OrdersPage
│   │   ├── components/       # Shared UI
│   │   ├── hooks/            # useWebSocket, useCart, useTelegram
│   │   ├── lib/              # API client, ws client, utils
│   │   ├── store/            # Zustand state (cart, auth, orders)
│   │   └── types/            # TypeScript interfaces
├── admin/           # Nuxt.js (B2B panel)
│   ├── pages/
│   │   ├── login.vue
│   │   ├── dashboard.vue       # Today's orders, stats
│   │   ├── orders/index.vue    # Kanban board (New | Preparing | Ready)
│   │   ├── menu/index.vue      # Categories list
│   │   ├── menu/[categoryId].vue # Items + modifiers
│   │   └── settings.vue        # Store info, staff
│   ├── components/
│   ├── composables/            # useWebSocket, useAuth, useApi
│   ├── lib/
│   └── layouts/default.vue     # Sidebar nav + header
└── docker-compose.yml
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22+, Echo framework |
| Database | PostgreSQL 16 |
| Customer app | React, TypeScript, Vite, Zustand, Tailwind CSS, shadcn/ui |
| Admin panel | Nuxt.js 3, TypeScript, Tailwind CSS, shadcn-vue |
| Real-time | WebSockets (native, via Echo) |
| Telegram | Bot API, Mini App SDK (`@telegram-apps/sdk-react`), Telegram Gateway |
| Dev environment | Docker Compose |

## Data Model

### stores

The central tenant entity.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| name | VARCHAR | Display name |
| code | VARCHAR | Unique short code for staff login (e.g., `barname01`) |
| slug | VARCHAR | URL-friendly identifier |
| description | TEXT | |
| address | VARCHAR | |
| phone | VARCHAR | |
| logo_url | VARCHAR | |
| telegram_group_chat_id | BIGINT | For bot notifications to store staff |
| subscription_tier | VARCHAR | `free`, `basic`, `premium` |
| subscription_expires_at | TIMESTAMP | |
| commission_rate | DECIMAL(5,2) | Percentage (e.g., 5.00) |
| is_active | BOOLEAN | |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

### store_staff

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| store_id | UUID | FK → stores |
| staff_code | VARCHAR | Login identifier within the store |
| name | VARCHAR | |
| password_hash | VARCHAR | bcrypt |
| role | VARCHAR | `owner`, `manager`, `barista` |
| is_active | BOOLEAN | |
| created_at | TIMESTAMP | |

### users

Telegram users who place orders.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| telegram_id | BIGINT | Unique, from Telegram |
| phone | VARCHAR | Verified via Telegram Gateway |
| first_name | VARCHAR | |
| last_name | VARCHAR | |
| username | VARCHAR | Telegram username |
| created_at | TIMESTAMP | |

### categories

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| store_id | UUID | FK → stores |
| name | VARCHAR | |
| sort_order | INT | |
| is_active | BOOLEAN | |

### items

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| category_id | UUID | FK → categories |
| store_id | UUID | FK → stores (denormalized for query efficiency) |
| name | VARCHAR | |
| description | TEXT | |
| base_price | BIGINT | Price in UZS (integer, no decimals) |
| image_url | VARCHAR | |
| is_available | BOOLEAN | |
| sort_order | INT | |

### modifier_groups

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| item_id | UUID | FK → items |
| store_id | UUID | FK → stores (denormalized) |
| name | VARCHAR | e.g., "Choose size", "Add extras" |
| selection_type | VARCHAR | `single`, `multiple` |
| is_required | BOOLEAN | |
| min_selections | INT | 0 for optional |
| max_selections | INT | |
| sort_order | INT | |

### modifiers

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| modifier_group_id | UUID | FK → modifier_groups |
| store_id | UUID | FK → stores (denormalized) |
| name | VARCHAR | e.g., "Large", "Extra shot" |
| price_adjustment | BIGINT | In UZS. Can be 0. |
| is_available | BOOLEAN | |
| sort_order | INT | |

### orders

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| order_number | INT | Sequential per store, human-readable (e.g., #42) |
| user_id | UUID | FK → users |
| store_id | UUID | FK → stores |
| status | VARCHAR | See order statuses below |
| total_price | BIGINT | UZS |
| payment_method | VARCHAR | `pay_at_pickup` (prototype default) |
| payment_status | VARCHAR | `pending`, `paid` |
| eta_minutes | INT | When user expects to arrive |
| rejection_reason | TEXT | If store rejects |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

**Order statuses:** `pending` → `accepted` → `preparing` → `ready` → `picked_up`

Also: `rejected` (store declines), `cancelled` (user cancels before acceptance).

### order_items

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| order_id | UUID | FK → orders |
| item_id | UUID | FK → items (reference, not for pricing) |
| item_name | VARCHAR | Snapshot |
| item_price | BIGINT | Snapshot — price at time of order |
| quantity | INT | |

### order_item_modifiers

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| order_item_id | UUID | FK → order_items |
| modifier_id | UUID | FK → modifiers (reference) |
| modifier_name | VARCHAR | Snapshot |
| price_adjustment | BIGINT | Snapshot |

### transactions

Commission tracking per order.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| order_id | UUID | FK → orders |
| store_id | UUID | FK → stores |
| order_total | BIGINT | |
| commission_rate | DECIMAL(5,2) | Rate at time of order |
| commission_amount | BIGINT | Calculated |
| created_at | TIMESTAMP | |

## API Routes

### Public (Customer Mini App)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/telegram` | Verify Telegram `initData`, issue JWT |
| GET | `/stores/:slug` | Store info |
| GET | `/stores/:slug/menu` | Full menu: categories → items → modifier groups → modifiers |
| POST | `/orders` | Create order |
| GET | `/orders/:id` | Order status |
| PUT | `/orders/:id/cancel` | User cancels (only if `pending`) |
| WS | `/ws` | WebSocket — authenticated, scoped to user's active orders |

### Admin (Store Staff — Nuxt Panel)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/admin/auth` | Login with `{ store_code, staff_code, password }` |
| GET | `/admin/store` | Store details (from JWT `store_id`) |
| PUT | `/admin/store` | Update store info |
| GET | `/admin/menu/categories` | List categories |
| POST | `/admin/menu/categories` | Create category |
| PUT | `/admin/menu/categories/:id` | Update category |
| DELETE | `/admin/menu/categories/:id` | Delete category |
| GET | `/admin/menu/categories/:id/items` | List items in category |
| POST | `/admin/menu/items` | Create item |
| PUT | `/admin/menu/items/:id` | Update item |
| DELETE | `/admin/menu/items/:id` | Delete item |
| POST | `/admin/menu/items/:id/modifier-groups` | Create modifier group |
| PUT | `/admin/menu/modifier-groups/:id` | Update modifier group |
| DELETE | `/admin/menu/modifier-groups/:id` | Delete modifier group |
| POST | `/admin/menu/modifier-groups/:id/modifiers` | Create modifier |
| PUT | `/admin/menu/modifiers/:id` | Update modifier |
| DELETE | `/admin/menu/modifiers/:id` | Delete modifier |
| GET | `/admin/orders` | List orders (filterable by status) |
| PUT | `/admin/orders/:id/status` | Update status (`accepted`, `preparing`, `ready`, `rejected`) |
| WS | `/admin/ws` | WebSocket — new orders, cancellations |

## Authentication

### User Auth (Telegram Mini App)

1. User interacts with Telegram bot → bot verifies phone via Telegram Gateway → stores phone in `users`
2. User opens Mini App → app sends Telegram `initData` to `POST /auth/telegram`
3. Server validates `initData` signature using bot token → issues JWT with `user_id` and `telegram_id`

### Store Staff Auth (Admin Panel)

1. Staff enters store code, staff code, and password on login page
2. `POST /admin/auth` validates all three fields → issues JWT with `store_id`, `staff_id`, `role`
3. All `/admin/*` routes require valid JWT; middleware extracts `store_id` for tenant scoping

## Real-time Communication

### WebSocket Protocol

**Client → Server:**
```json
{"type": "subscribe", "channel": "order:123"}
{"type": "subscribe", "channel": "store:orders"}
```

**Server → User:**
```json
{"type": "order:status", "order_id": "uuid", "status": "preparing", "updated_at": "..."}
{"type": "order:rejected", "order_id": "uuid", "reason": "We're closing soon"}
```

**Server → Store Admin:**
```json
{"type": "order:new", "order": { "...full order object..." }}
{"type": "order:cancelled", "order_id": "uuid"}
```

Auto-reconnect with exponential backoff on disconnect.

### Telegram Bot Notifications

**To user (bot DM):**
- Order accepted: "Your order #42 at BarName has been accepted! Preparing now."
- Order preparing: "Your order #42 is being prepared."
- Order ready: "Your order #42 is ready for pickup!"
- Order rejected: "Sorry, BarName couldn't accept your order. Reason: ..."

**To store (group chat):**
- New order: "New order #42! 2x Mojito (Large), 1x Nachos. Customer arrives in ~15 min. Total: 120,000 UZS"
- Order cancelled: "Order #42 was cancelled by the customer."

Both WebSocket and bot notifications fire simultaneously on every order status change.

## Multi-tenancy

- **Row-level isolation** via `store_id` foreign key on all tenant-scoped tables
- Admin JWT embeds `store_id`; middleware injects into request context
- Repository layer always filters by `store_id`
- Users table is shared — a user can order from multiple stores

### Subscription & Commissions (schema only for prototype)

- `stores` table includes `subscription_tier`, `subscription_expires_at`, `commission_rate`
- `transactions` table logs commission per order automatically
- No billing UI in prototype — data collection only

### Store Onboarding (prototype)

Manual — seed script or internal endpoint creates store + first admin account.

## User Flows

### Customer Ordering Flow

1. User opens Telegram bot → bot sends Mini App button for a store
2. Mini App opens → auto-authenticated via Telegram `initData`
3. Browse menu by category → tap item → select modifiers → add to cart
4. Cart page → set "I'll arrive in X minutes" → place order
5. Order page — real-time status: Pending → Accepted → Preparing → Ready
6. Bot sends message at each status change

### Store Order Management Flow

1. Staff logs in to admin panel (store code + staff code + password)
2. Dashboard shows today's active orders and stats
3. Orders page — kanban board: **New | Preparing | Ready**
4. New order arrives (WebSocket + sound alert + Telegram group message)
5. Staff taps Accept/Reject → moves through Preparing → Ready
6. Each transition notifies the customer via WebSocket + bot

## Development Environment

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: xpressgo
      POSTGRES_USER: xpressgo
      POSTGRES_PASSWORD: xpressgo
    ports:
      - "5432:5432"

  server:
    build: ./server
    depends_on:
      - postgres
    environment:
      DATABASE_URL: postgres://xpressgo:xpressgo@postgres:5432/xpressgo?sslmode=disable
      TELEGRAM_BOT_TOKEN: ${TELEGRAM_BOT_TOKEN}
      JWT_SECRET: ${JWT_SECRET}
    ports:
      - "8080:8080"

  web:
    build: ./web
    ports:
      - "5173:5173"

  admin:
    build: ./admin
    ports:
      - "3000:3000"
```

**Telegram dev setup:** Use ngrok or similar to expose local server for Telegram webhook callbacks.

## Deferred (Post-Prototype)

- Payme payment integration
- Production deployment
- Global admin panel (Xpressgo internal management)
- Self-service store onboarding
- Billing dashboard (subscription management, commission reports)
- Order analytics for stores
- Image upload for menu items
- Push notifications beyond Telegram
