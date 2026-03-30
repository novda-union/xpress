# Xpressgo — Complete UI/UX Design System

## Design Philosophy

**Mini App:** Feels native inside Telegram. Adapts to Telegram's dark/light theme automatically. Warm, food-forward brand accent. Every tap feels immediate and satisfying.

**Admin Panel:** Professional, distraction-free B2B tool. Clean enough to use on a busy bar shift. Data is always visible, never buried.

---

## Part 1: Mini App (React + Tailwind + shadcn/ui)

### 1.1 Color System

The mini app builds on **Telegram's CSS variables** so it automatically adapts to the user's Telegram theme (dark or light). Brand colors overlay on top.

```css
/* Telegram CSS variables — already provided by the WebApp SDK */
--tg-theme-bg-color           /* primary background */
--tg-theme-secondary-bg-color /* card / surface background */
--tg-theme-text-color         /* primary text */
--tg-theme-hint-color         /* muted / secondary text */
--tg-theme-link-color         /* links */
--tg-theme-button-color       /* Telegram's accent button bg */
--tg-theme-button-text-color  /* Telegram's accent button text */

/* Xpressgo brand layer — override in :root */
--xp-brand:        #FF5A00;   /* vibrant orange — primary CTAs, active tabs, accents */
--xp-brand-muted:  #FF5A0020; /* orange tint for backgrounds / selection highlights */
--xp-success:      #22C55E;   /* order ready, in-stock */
--xp-warning:      #F59E0B;   /* preparing state */
--xp-error:        #EF4444;   /* rejection, errors */
--xp-overlay:      rgba(0, 0, 0, 0.5); /* map bottom sheet backdrop */

/* Surface cards — adapt to theme */
--xp-card-bg:     var(--tg-theme-secondary-bg-color);
--xp-border:      rgba(128, 128, 128, 0.15);
```

**Status color map (order states):**

| State | Color | Usage |
|-------|-------|-------|
| pending | `#94A3B8` slate | Neutral waiting |
| accepted | `#3B82F6` blue | Confirmed |
| preparing | `#F59E0B` amber | In progress |
| ready | `#22C55E` green | Pickup time |
| picked_up | `#6366F1` indigo | Complete |
| rejected | `#EF4444` red | Failed |
| cancelled | `#94A3B8` slate | Cancelled |

### 1.2 Typography

**Fonts:** `Inter Variable` for body and controls, with `Geist Variable` available for headings and hero accents.

```css
@import "@fontsource-variable/inter";
@import "@fontsource-variable/geist";

font-family: 'Inter Variable', -apple-system, BlinkMacSystemFont, sans-serif;
```

- `Inter Variable` remains the default mini app font
- `Geist Variable` may be used for prominent headings and display copy
- Telegram theme variables still remain the primary adaptation layer for dark and light appearance

**Type Scale:**

| Name | Size | Weight | Line Height | Use |
|------|------|--------|-------------|-----|
| `display` | 28px | 700 | 1.2 | Store name on menu page |
| `heading` | 22px | 700 | 1.25 | Page titles |
| `title` | 18px | 600 | 1.35 | Section headings, item name |
| `body` | 15px | 400 | 1.5 | Descriptions, body text |
| `body-med` | 15px | 500 | 1.5 | Labels, navigation |
| `small` | 13px | 400 | 1.4 | Metadata, distance, timestamps |
| `caption` | 11px | 500 | 1.3 | Badges, category pills |

**Rule:** Never go below 13px. Body text minimum 15px on mobile.

### 1.3 Spacing Scale

Based on 4px grid:

```
4px   — xs  (icon internal padding)
8px   — sm  (between tight elements)
12px  — md  (card internal padding top/bottom)
16px  — lg  (standard horizontal page padding, card padding)
20px  — xl  (section gaps)
24px  — 2xl (between cards)
32px  — 3xl (large section spacing)
```

Page horizontal padding: `16px` on all screens.

### 1.4 Border Radius Scale

```
6px   — sm  (tags, small badges)
12px  — md  (input fields, small cards)
16px  — lg  (item cards, standard cards)
20px  — xl  (bottom sheet top corners, large modals)
9999px — full (pills, circular elements)
```

### 1.5 Shadows

Keep shadows subtle — they work on both dark and light Telegram themes.

```css
--shadow-sm:  0 1px 3px rgba(0,0,0,0.08), 0 1px 2px rgba(0,0,0,0.06);
--shadow-md:  0 4px 12px rgba(0,0,0,0.12), 0 2px 4px rgba(0,0,0,0.08);
--shadow-lg:  0 8px 32px rgba(0,0,0,0.18), 0 4px 8px rgba(0,0,0,0.10);
--shadow-map-marker: 0 2px 8px rgba(0,0,0,0.25);
```

### 1.6 Animation System

**Rule:** ease-out for entering elements, ease-in for exiting. Never use `linear` for UI.

```css
/* Micro-interactions: buttons, tabs, toggles */
--anim-fast:    150ms ease-out;

/* Page transitions, card appearances */
--anim-normal:  220ms ease-out;

/* Bottom sheets, modals, large elements */
--anim-slow:    300ms cubic-bezier(0.32, 0.72, 0, 1);  /* iOS spring-like */

/* Map marker pop-in */
--anim-spring:  400ms cubic-bezier(0.34, 1.56, 0.64, 1); /* overshoot spring */
```

```css
/* Always check reduced motion */
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

### 1.7 Screen-by-Screen Specs

---

#### Phone Auth Gate

**Layout:** Full screen, vertically centered content.

```
┌─────────────────────────────────┐
│                                 │
│         [Xpressgo Logo]         │  — 48px SVG logo, centered
│                                 │
│      Order without the wait     │  — heading, 22px bold
│  We need your phone to identify │  — body, 15px, hint color
│   you and keep your orders safe │
│                                 │
│  ┌───────────────────────────┐  │
│  │  📱  Share Phone Number   │  │  — 54px height, brand orange, full width
│  └───────────────────────────┘  │    rounded-xl, Inter 600 16px
│                                 │
│   Your number is only used for  │  — caption, 13px, hint color
│         order tracking          │
│                                 │
└─────────────────────────────────┘
```

- **Error state (Deny):** Red-tinted card below the button: `"Phone number is required to continue"` with a retry button underneath. No shake animation — just a smooth fade-in of the error card.
- **Loading state:** Button shows spinner, disabled. `opacity-70 cursor-not-allowed`.
- **Button icon:** Use `PhoneCall` from Lucide, 18px, left of text.

---

#### Home Screen — Map/List Toggle

**Current implementation note:**

- the discovery view toggle now floats near the bottom center for better thumb reach
- the top bar remains focused on discovery context rather than carrying the toggle itself
- map and list views still share one discovery state and only the presentation swaps

**Top Bar:**
```
┌─────────────────────────────────┐
│  Xpressgo                       │  — logo and discovery context
└─────────────────────────────────┘
```

Toggle component:
- Container: floating glass panel with blur, rounded-full border, and elevated shadow
- Each segment: icon-first control with minimum `44px` touch target
- Active: sliding `var(--xp-brand)` indicator with white active icon color
- Transition: spring-like indicator movement plus eased color change
- Icons: `Map` and `LayoutList` from Lucide

**Map View:**
- MapLibre fills the remaining screen height below the top bar
- No padding — edge-to-edge
- Maptiler "Dataviz Dark" style when Telegram dark mode, "Streets" style for light mode
- User location dot: blue circle with white border + pulsing ring animation
- Attribution text: bottom right, 10px, semi-transparent

**List View:**
- Category tabs: sticky below top bar, horizontal scroll, no scrollbar visible (`scrollbar-none`)
- Content: scrollable list of branch cards
- Both views use the same top bar — just the content area swaps

---

#### Map Branch Markers

**Unselected state:**
- 44×44px circle
- Store `logo_url` image, `object-cover`, white border `3px`
- `box-shadow: var(--shadow-map-marker)`
- Pop-in on load: scale from 0.3 → 1.0 with `var(--anim-spring)`, staggered 50ms per marker

**Selected state:**
- Scale to 1.15× with `var(--anim-fast)` spring
- White border increases to `4px`
- Drops a slightly larger shadow

**Marker HTML (rendered via MapLibre DOM markers):**
```html
<div class="marker [unselected|selected]">
  <img src="..." alt="Store logo" />
</div>
```

---

#### Branch Bottom Sheet Card

Triggered by tapping a marker. Slides up from bottom.

**Animation:**
- Enter: `translateY(100%) → translateY(0)` in `300ms cubic-bezier(0.32, 0.72, 0, 1)`
- Exit: `translateY(0) → translateY(100%)` in `220ms ease-in`
- Backdrop: `rgba(0,0,0,0.5)` fades in simultaneously over the map

**Layout:**
```
┌─────────────────────────────────┐
│           ─────              │  — drag handle: 4×32px, rounded, hint color, centered, mt-3
│                               │
│  [──────── Banner Image ──────]│  — 160px height, full width, no border-radius
│                               │
│  Xpressgo Bar          ×      │  — store name (title 18/600) + close icon (X, 20px)
│  Branch - Chilonzor           │  — branch name (small, hint color)
│  📍 Amir Temur St, 15         │  — MapPin icon (12px hint color) + address
│                               │
│  ┌────────────────────────────┐│
│  │ [Item] [Item] [Item] [Item]││  — horizontal scroll, no scrollbar
│  └────────────────────────────┘│    each item: 100×130px card (image + name + price)
│                               │
│  ┌────────────────────────────┐│
│  │      See Full Menu  →      ││  — brand orange button, full width, 52px height
│  └────────────────────────────┘│    rounded-xl, Inter 600
└─────────────────────────────────┘
```

**Menu carousel item card:**
- `100×130px` total, `rounded-xl`, `bg-[var(--xp-card-bg)]`
- Image: `100×80px` top, `rounded-t-xl`, `object-cover`
- Name: 12px, 500, 2-line clamp, `8px` horizontal padding
- Price: 12px, 600, brand orange, `8px` horizontal padding, `4px` bottom padding
- `cursor-pointer`, hover: `opacity-90 scale-[1.02]` in `150ms`

---

#### List View — Category Tabs

```
[ All ]  [ Bars ]  [ Cafes ]  [ Coffee ]  [ Restaurants ]  [ Fast Food ]
```

- Container: `px-4 py-3`, `overflow-x-auto scrollbar-none`
- Tab pill: `px-4 h-9 rounded-full text-[13px] font-medium whitespace-nowrap`
- Active: `bg-[var(--xp-brand)] text-white`
- Inactive: `bg-[var(--xp-card-bg)] text-[var(--tg-theme-hint-color)]`
- Gap between pills: `8px`
- Transition: background + color `150ms ease-out`

---

#### List View — Branch Card

```
┌──────────────────────────────────────┐
│ ┌────────┐  Xpressgo Bar            │
│ │  Logo  │  Branch - Chilonzor      │
│ │ 72×72  │  📍 Amir Temur St, 15   │
│ └────────┘  [Bar]           1.2 km  │
└──────────────────────────────────────┘
```

- Container: `rounded-2xl bg-[var(--xp-card-bg)] p-4` with `var(--shadow-sm)`
- Logo: `72×72px rounded-xl object-cover` (no circular — keeps brand feel)
- Store name: `15px 600`
- Branch name: `13px 400 hint-color`
- Address: `13px hint-color` with `MapPin` icon `12px`
- Category badge: `px-2 py-0.5 rounded-full text-[11px] font-medium` — color per category:
  - Bar: `bg-purple-500/15 text-purple-600`
  - Cafe: `bg-amber-500/15 text-amber-600`
  - Coffee: `bg-brown-500/15 text-orange-700` (use `bg-orange-900/10`)
  - Restaurant: `bg-green-500/15 text-green-600`
  - Fast Food: `bg-red-500/15 text-red-600`
- Distance: `13px 500 hint-color` right-aligned
- Hover: `opacity-95` + `translateY(-1px)` in `150ms`
- `cursor-pointer` always

---

#### Branch Menu Page

**Header:**
- Fixed top, `48px` height
- Back arrow (ChevronLeft, 24px) left
- Store name centered, `18px 600`
- Cart icon right (ShoppingCart, 22px) with badge showing item count
- Background: `var(--tg-theme-bg-color)` with `backdrop-blur-sm` — glass effect as user scrolls

**Store hero section:**
- Banner image: full width, `200px` height, `object-cover`, no border-radius
- Store name overlay at bottom: gradient from transparent to `rgba(0,0,0,0.6)`, white text `22px 700`

**Category tabs:** Same pill style as list view. Sticky below header.

**Item Grid:** 2 columns, `gap-3`, `px-4`

---

#### Item Card (Menu Grid)

```
┌─────────────────┐
│                 │
│   [Image 100%]  │  — 140px height, rounded-t-2xl, object-cover
│                 │
│ Mojito          │  — 15px 600, mt-2, mx-3
│ Fresh mint,     │  — 13px hint-color, 2-line clamp, mx-3
│ lime & rum      │
│                 │
│ 32,000 UZS  [+] │  — price 14px 600 brand-orange, + button right
└─────────────────┘
```

- Container: `rounded-2xl bg-[var(--xp-card-bg)]` with `var(--shadow-sm)`
- `+` button: `32×32px` circle, `bg-[var(--xp-brand)] text-white`, bottom-right `mr-3 mb-3`
- `+` tap: scale `0.9 → 1.0` spring in `150ms` (haptic-like)
- Unavailable item: banner image is `grayscale(1)`, overlay `"Unavailable"` badge top-right

---

#### Item Detail Page (`/item/:id`)

**Layout:** Scrollable content with sticky bottom bar.

```
[← Back]                     [Cart 2]   ← sticky transparent header (fills on scroll)
┌───────────────────────────────────┐
│                                   │
│          [Hero Image]             │  — full width, 280px, no top border-radius
│                                   │
└───────────────────────────────────┘
│ Mojito                            │  — 22px 700, mt-4, px-4
│ 32,000 UZS                        │  — 20px 600 brand-orange, px-4
│ Fresh mint, premium rum,          │  — 15px hint-color, mt-2, px-4
│ lime juice and soda water.        │
│                                   │
│ ── Choose Size ──────────────────  │  — modifier group header: 13px uppercase hint, px-4
│ ┌──────────────────────────────┐  │
│ │ ○  Regular     +0 UZS       │  │  — radio item: 52px height, px-4
│ │ ●  Large      +8,000 UZS    │  │    active: brand-orange radio dot
│ │ ○  XL         +15,000 UZS   │  │
│ └──────────────────────────────┘  │
│                                   │
│ ── Add Extras ─────────────────── │
│ ┌──────────────────────────────┐  │
│ │ ☐  Extra shot  +5,000 UZS  │  │  — checkbox item
│ └──────────────────────────────┘  │
│                                   │
│                                   │  — bottom padding: 100px (space for sticky bar)
└───────────────────────────────────┘

┌───────────────────────────────────┐  ← sticky bottom bar
│  ─ 1 +              Add to Cart  │  — quantity selector left, CTA right
│                      45,000 UZS  │    price updates live
└───────────────────────────────────┘
```

**Sticky bottom bar:**
- `h-[80px]` + `pb-safe` (safe area inset)
- `bg-[var(--tg-theme-bg-color)]` with `border-t border-[var(--xp-border)]`
- Quantity selector: `─ [n] +` with `36×36px` buttons, `24px` number
- CTA button: brand orange, `rounded-xl`, `flex-1 ml-4`, `52px` height, shows updated total
- Add to cart: brief scale animation on tap `1.0 → 0.96 → 1.0` in `150ms`

**Modifier group styling:**
- Section header: `text-[11px] font-semibold uppercase tracking-wide hint-color` with full-width divider
- Radio/checkbox items: `px-4 py-3.5 min-h-[52px]` — touch target safe
- Active radio: custom orange dot (don't use browser default)
- Hover/tap: `bg-[var(--xp-brand-muted)]` tint in `100ms`

---

#### Cart Page

```
┌─────────────────────────────────────┐
│  ←  Your Cart            (2 items)  │
├─────────────────────────────────────┤
│                                     │
│  [Item image 48px]  Mojito          │
│                     Large           │
│                     − 1 +   32,000  │
│  ────────────────────────────────   │
│  [Item image 48px]  Nachos          │
│                     − 2 +   24,000  │
│                                     │
│  ────────────────────────────────── │
│  Arrive in                          │
│  [5 min][10 min][15 min][20 min]... │  — horizontal scroll ETA pills
│                                     │
│  ────────────────────────────────── │
│  Subtotal                 56,000 ₩  │
│  ────────────────────────────────── │
│  ┌──────────────────────────────┐   │
│  │   Place Order — 56,000 UZS  │   │  — brand orange, 54px, full width, rounded-xl
│  └──────────────────────────────┘   │
└─────────────────────────────────────┘
```

- Item row height: `72px`, swipe-to-delete (optional for prototype)
- Item thumbnail: `48×48px rounded-xl object-cover`
- Quantity controls: `−` and `+` are `32×32px` circles, `bg-[var(--xp-card-bg)]`
- ETA pills: same style as category tabs but with clock icon before text
- Subtotal: `16px 600`, right-aligned
- CTA shows `opacity-50 cursor-not-allowed` when no ETA selected

---

#### Order Tracking Page

```
┌─────────────────────────────────────┐
│  ←  Order #42                       │
│                                     │
│         [Status Badge]              │  — large, centered, colored per state
│       "Being Prepared"              │  — 18px 600 text-color
│     Your order is on its way!       │  — 14px hint-color
│                                     │
│  ○────●────○────○────○              │  — progress bar: 5 steps
│  Pending Accepted Prep  Ready Picked│    active step: brand orange circle + fill
│                                     │
│  ────────────────────────────────── │
│  Xpressgo Bar · Branch Chilonzor    │
│  ────────────────────────────────── │
│  2× Mojito (Large)        64,000    │
│  1× Nachos                12,000    │
│  ────────────────────────────────── │
│  Total                    76,000    │
│                                     │
│  ────────────────────────────────── │
│  ┌──────────────────────────────┐   │
│  │       Cancel Order           │   │  — ghost button, only shown when pending
│  └──────────────────────────────┘   │
└─────────────────────────────────────┘
```

**Progress bar:**
- 5 circles connected by a line
- Completed steps: filled brand orange circle + filled line segment
- Current step: brand orange circle with pulsing ring `animate-ping` at `0.5 opacity`
- Future steps: `var(--xp-border)` colored circle
- Step label: `10px` below each circle
- Line fill animates from left on state change: `width: 0% → 100%` in `600ms ease-out`

**Status badge:**
- `px-4 py-2 rounded-full text-[15px] font-semibold` with color from status color map
- Background: `{color}/15` tint, text: full color
- Scale bounce on status change: `1.0 → 1.1 → 1.0` in `300ms`

---

#### Order History Page

- List of past order cards
- Each card: order number, store + branch name, status badge, total, date
- Card height: `80px`, `px-4 py-4`
- Empty state: centered illustration placeholder + "No orders yet" text
- `cursor-pointer`, tap → navigates to order detail

---

### 1.8 Global Mini App Rules

- **Safe area:** Always `padding-bottom: env(safe-area-inset-bottom)` on bottom bars
- **No horizontal scroll:** `overflow-x: hidden` on body
- **Touch targets:** All interactive elements minimum `44×44px`
- **Disabled states:** `opacity-50 cursor-not-allowed pointer-events-none`
- **Images:** Always `object-cover` with explicit dimensions to prevent layout shift
- **Icons:** Lucide React exclusively, consistent `w-5 h-5` (20px) default
- **Loading skeletons:** Pulse animation `animate-pulse bg-[var(--xp-border)]` — match exact element shape

---

## Part 2: Admin Panel (Nuxt.js 3 + Tailwind + shadcn-vue)

**Implementation note:**

- the admin panel now has a concrete shared component layer in `admin/components/ui`
- pages and domain components should compose shared sidebar, card, button, input, select, sheet, avatar, and tooltip primitives
- avoid reintroducing bespoke one-off UI patterns when an existing shared primitive already fits

### 2.1 Color System

```css
--background:                 oklch(1 0 0);
--foreground:                 oklch(0.145 0 0);
--card:                       oklch(1 0 0);
--card-foreground:            oklch(0.145 0 0);
--primary:                    oklch(0.585 0.226 264);
--primary-foreground:         oklch(0.985 0 0);
--secondary:                  oklch(0.97 0 0);
--secondary-foreground:       oklch(0.205 0 0);
--muted:                      oklch(0.97 0 0);
--muted-foreground:           oklch(0.556 0 0);
--border:                     oklch(0.922 0 0);
--input:                      oklch(0.922 0 0);
--ring:                       oklch(0.585 0.226 264);
--sidebar:                    oklch(0.985 0 0);
--sidebar-foreground:         oklch(0.145 0 0);
--sidebar-primary:            oklch(0.585 0.226 264);
--sidebar-primary-foreground: oklch(0.985 0 0);
--sidebar-accent:             oklch(0.95 0.02 264);
--sidebar-accent-foreground:  oklch(0.585 0.226 264);
```

These tokens are consumed through Tailwind and shadcn-vue primitives rather than bespoke `--admin-*` utility classes.

### 2.2 Typography

**Font:** `Plus Jakarta Sans` — friendly, modern, perfect for SaaS/B2B.

```css
@import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap');
font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, sans-serif;
```

**Type Scale:**

| Name | Size | Weight | Use |
|------|------|--------|-----|
| `page-title` | 24px | 700 | Page headings |
| `section-title` | 18px | 600 | Card titles, section headings |
| `label` | 14px | 600 | Form labels, table headers |
| `body` | 14px | 400 | Body text, table cells |
| `small` | 13px | 400 | Metadata, timestamps |
| `badge` | 12px | 600 | Status badges, role pills |
| `stat-number` | 32px | 700 | Dashboard KPI numbers |
| `stat-label` | 13px | 500 | KPI label below number |

### 2.3 Layout

**Sidebar:** Fixed left, `240px` width, `100vh` height.

```
┌────────┬──────────────────────────────────┐
│        │  Page Header                     │
│  Side  │  ───────────────────────────     │
│  bar   │                                  │
│ 240px  │  Page Content                    │
│        │                                  │
│        │                                  │
└────────┴──────────────────────────────────┘
```

Content area: `ml-[240px]`, `min-h-screen bg-[var(--admin-bg)]`.

Page content padding: `px-6 py-6` (desktop), `px-4 py-4` (tablet).

### 2.4 Component Specs

---

#### Sidebar

```
┌────────────────────────┐
│  [Logo] Xpressgo       │  — 64px height header
│                        │
│  [Branch Switcher ▾]   │  — Director only, full width, 40px
│  ───────────────────   │
│  🏠  Dashboard         │  — nav item, 40px height
│  📋  Orders       (3)  │  — with badge
│  🍔  Menu              │
│  🏪  Branches          │
│  👥  Staff             │
│  ⚙️  Settings          │
│                        │
│  ───────────────────   │
│  [Avatar] John D.      │  — user info, 48px height
│  Manager  ·  Log out   │
└────────────────────────┘
```

- Background: `var(--admin-surface)`, `border-r border-[var(--admin-border)]`
- Nav item: `flex items-center gap-3 px-4 h-10 rounded-lg mx-2 text-[14px] font-medium`
- Inactive: `text-[var(--admin-text-muted)] hover:bg-[var(--admin-surface-2)] hover:text-[var(--admin-text)]`
- Active: `bg-[var(--admin-accent-bg)] text-[var(--admin-accent)]`
- Icon: `w-5 h-5` (20px) Lucide, consistent
- Badge: `ml-auto px-2 py-0.5 rounded-full bg-[var(--admin-error)] text-white text-[11px] font-semibold`
- Transition: `150ms ease-out` on hover

**Branch Switcher (Director):**
- `px-4 py-2 rounded-lg bg-[var(--admin-surface-2)] border border-[var(--admin-border)]`
- Shows current branch name or "All Branches"
- ChevronDown icon right
- Dropdown: `z-50 shadow-lg rounded-xl border` with branch list + "All Branches" option

---

#### Page Header

```
┌──────────────────────────────────────────────────┐
│  Dashboard              [Add Branch] [Export]    │
│  Sunday, March 29 · All Branches                 │
└──────────────────────────────────────────────────┘
```

- `h-16 px-6 flex items-center justify-between`
- `border-b border-[var(--admin-border)] bg-[var(--admin-surface)]`
- Title: `20px 700`
- Subtitle: `13px text-muted`
- Actions: right-aligned, `gap-3`

---

#### Login Page

```
┌──────────────────────────────────┐
│           [Logo 48px]            │
│          Xpressgo Admin          │  — 24px 700
│  Welcome back, sign in to        │  — 14px muted
│  manage your store               │
│                                  │
│  Store Code                      │
│  ┌──────────────────────────┐    │
│  │  e.g. demobar01          │    │  — 40px height, rounded-lg
│  └──────────────────────────┘    │
│                                  │
│  Staff Code                      │
│  ┌──────────────────────────┐    │
│  │                          │    │
│  └──────────────────────────┘    │
│                                  │
│  Password                        │
│  ┌──────────────────────────┐    │
│  │  ········         [👁]  │    │  — Eye icon toggle
│  └──────────────────────────┘    │
│                                  │
│  ┌──────────────────────────┐    │
│  │         Sign In          │    │  — indigo, 42px, rounded-lg, full width
│  └──────────────────────────┘    │
└──────────────────────────────────┘
```

- Centered card: `max-w-sm mx-auto mt-20 p-8 bg-white rounded-2xl shadow-lg`
- Form fields: `h-10 rounded-lg border border-[var(--admin-border)]`, focus ring indigo
- Error: red border + small red text below field
- Submit loading: spinner inside button, disabled state

---

#### Dashboard — KPI Cards (Bento Layout)

```
┌──────────┬──────────┬──────────┬──────────┐
│  Orders  │ Revenue  │ Pending  │  Ready   │
│  Today   │  Today   │  Now     │  Pickup  │
│          │          │          │          │
│   142    │ 4.2M UZS │    8     │    3     │
│   +12%   │  +5.4%   │          │          │
└──────────┴──────────┴──────────┴──────────┘
```

- Grid: `grid grid-cols-4 gap-4`
- Card: `bg-[var(--admin-surface)] rounded-2xl p-5 border border-[var(--admin-border)]`
- Stat number: `32px 700 text-[var(--admin-text)]`
- Change badge: green for positive, red for negative — `text-[12px] font-semibold`
- Icon: top-right of card, `32×32px` circle background with accent tint, `w-5 h-5` icon inside
- Hover: `shadow-md translateY(-1px)` in `150ms`

---

#### Orders Page — Kanban Board

```
┌────────────────────┬──────────────────┬──────────────────┐
│  New Orders (5)    │  Preparing (3)   │  Ready (2)       │
│ ──────────────     │ ─────────────    │ ───────────────  │
│  [Order Card]      │  [Order Card]    │  [Order Card]    │
│  [Order Card]      │  [Order Card]    │  [Order Card]    │
│  [Order Card]      │  [Order Card]    │                  │
│  [Order Card]      │                  │                  │
│  [Order Card]      │                  │                  │
└────────────────────┴──────────────────┴──────────────────┘
```

**Column styling:**
- New: header `text-[var(--admin-new)]`, top border `border-t-2 border-[var(--admin-new)]`
- Preparing: header `text-[var(--admin-preparing)]`, `border-[var(--admin-preparing)]`
- Ready: header `text-[var(--admin-ready)]`, `border-[var(--admin-ready)]`
- Column bg: `bg-[var(--admin-surface-2)] rounded-2xl p-3`
- Count badge: colored pill matching column

**Order Card:**
```
┌─────────────────────────────────────┐
│  #42  •  3 items        2 min ago   │  — order number 14px 700, time small muted
│  ─────────────────────────────────  │
│  2× Mojito (Large)                  │
│  1× Nachos                          │
│                                     │
│  76,000 UZS              15 min ETA │  — total 14px 600, ETA small
│  ─────────────────────────────────  │
│  [✓ Accept]          [✗ Reject]     │  — New column only
│  OR: [Preparing ▶]                  │  — Accepted → Preparing
│  OR: [Mark Ready ✓]                 │  — Preparing → Ready
│  OR: [Picked Up ✓]                  │  — Ready column
└─────────────────────────────────────┘
```

- Card: `bg-[var(--admin-surface)] rounded-xl p-4 border border-[var(--admin-border)]`
- New order card: left border `border-l-4 border-[var(--admin-new)]` + subtle `shadow-md`
- New order animation: slide in from top `translateY(-8px) opacity-0 → 0 opacity-1` in `300ms`
- Accept button: `bg-[var(--admin-success)] text-white rounded-lg h-9 text-[13px] font-semibold`
- Reject button: `border border-[var(--admin-error)] text-[var(--admin-error)] rounded-lg h-9 text-[13px]`
- Single action buttons (Preparing/Ready/PickedUp): full-width, colored per next state

---

#### Menu Management Page

**Category List:**
- Left panel `w-64`: scrollable category list
- Each category: `flex items-center justify-between px-4 h-12 rounded-lg cursor-pointer`
- Active: `bg-[var(--admin-accent-bg)] text-[var(--admin-accent)]`
- `+` Add Category: dashed border button at bottom

**Item List (right panel):**
- Grid `grid-cols-3 gap-4` or table view (toggleable)
- Item card: image 100%, `rounded-xl`, name + price + availability toggle
- Availability toggle: shadcn Switch, green when available
- Modifier groups: expandable accordion below item card
- Add Item: `+` floating button bottom-right

---

#### Branches Page

**Branch list:** Cards grid `grid-cols-2 gap-4`

**Branch card:**
```
┌─────────────────────────────────────┐
│  [Branch thumbnail 100%×120px]      │
│  Demo Bar - Chilonzor               │  — 16px 600
│  📍 Amir Temur St, 15               │  — 13px muted with icon
│  ● Active          8 Staff          │  — status dot + staff count
│  ─────────────────────────────────  │
│  [Edit]                [Deactivate] │
└─────────────────────────────────────┘
```

**Create/Edit Branch Form:**
- Slide-over panel (right side, `max-w-lg`) — not a modal
- Map pin picker: `320px` height MapLibre map embedded in form
- Instruction: `"Click on the map to set the branch location"`
- Lat/Lng fields: auto-filled, read-only, `text-[var(--admin-text-muted)]`
- Toggle active: shadcn Switch

---

#### Staff Management Page

**Staff table:**
| Name | Staff Code | Role | Branch | Status | Actions |
|------|-----------|------|--------|--------|---------|

- Role badge colors: director = purple, manager = indigo, barista = cyan
- Status: green dot + "Active" or grey dot + "Inactive"
- Row hover: `bg-[var(--admin-surface-2)]`
- Actions: Edit (pencil icon), Deactivate (slash icon) — icon buttons `36×36px`

**Add Staff form:**
- Slide-over panel, same as branches
- Role selector: radio card group (3 options with description)
- Branch selector: searchable dropdown (Director only)
- Password: generated suggestion + manual input option

---

#### Settings Page

**Two tab sections:**
- "Store Settings" (Director) and "Branch Settings" (Director + Manager)
- shadcn Tabs at top

**Settings form layout:**
- `max-w-2xl`
- Section headers: `16px 600` with separator line
- Input groups: label + input, `16px` gap between groups
- Save button: sticky at bottom of form, indigo

---

### 2.5 Admin Global Rules

- **Table density:** Comfortable (not compact) — `48px` row height minimum
- **Z-index scale:**
  - `10` — sticky headers, fixed bars
  - `20` — dropdowns, popovers
  - `30` — slide-over panels
  - `40` — modals / dialogs
  - `50` — toasts / notifications
- **Toasts:** Top-right, max-width `360px`, slide in from right, auto-dismiss `4s`
- **Empty states:** Centered in container, illustration icon `48px`, message + CTA
- **Loading states:** Full-width skeleton pulse on initial load; spinner on mutations
- **Transitions:** `150ms ease-out` on all hover states; `220ms ease-out` for panels

---

## Part 3: Shared Icons

Both apps use **Lucide** exclusively.

| Concept | Icon Name |
|---------|-----------|
| Menu/Food | `UtensilsCrossed` |
| Location | `MapPin` |
| Map | `Map` |
| List | `LayoutList` |
| Orders | `ClipboardList` |
| Cart | `ShoppingCart` |
| Staff | `Users` |
| Branch | `Store` |
| Settings | `Settings` |
| Dashboard | `LayoutDashboard` |
| Phone | `Phone` |
| Back | `ChevronLeft` |
| Close | `X` |
| Add | `Plus` |
| Delete | `Trash2` |
| Edit | `Pencil` |
| Accept | `Check` |
| Reject | `X` |
| Eye toggle | `Eye` / `EyeOff` |
| Logout | `LogOut` |
| Director role | `Crown` |
| Manager role | `Shield` |
| Barista role | `Coffee` |

---

## Part 4: Animation Checklist

- [ ] Map markers: spring scale pop-in on load, staggered 50ms
- [ ] Bottom sheet: translateY slide-up 300ms cubic-bezier spring
- [ ] Category tabs: 150ms active fill
- [ ] Item card `+` button: 150ms scale tap feedback
- [ ] Order status progress bar: 600ms segment fill on status change
- [ ] Status badge: scale bounce 300ms on change
- [ ] Kanban new order: slide-in from top 300ms
- [ ] Admin slide-over panels: translateX 250ms ease-out
- [ ] Sidebar nav active state: 150ms bg fill
- [ ] Toast: slide-in from right 220ms
- [ ] All: `prefers-reduced-motion` disables all transitions
