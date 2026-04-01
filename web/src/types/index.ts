export type StoreCategory = 'bar' | 'cafe' | 'restaurant' | 'coffee' | 'fastfood'

export interface Store {
  id: string
  name: string
  slug: string
  category: StoreCategory
  description: string
  address: string
  phone: string
  logo_url: string
}

export interface Branch {
  id: string
  store_id: string
  name: string
  address: string
  lat?: number | null
  lng?: number | null
  banner_image_url: string
  telegram_group_chat_id?: number | null
  is_active: boolean
}

export interface Category {
  id: string
  store_id: string
  branch_id: string
  name: string
  sort_order: number
  is_active: boolean
}

export interface Item {
  id: string
  category_id: string
  store_id: string
  branch_id: string
  name: string
  description: string
  base_price: number
  image_url: string
  is_available: boolean
  sort_order: number
}

export interface ModifierGroup {
  id: string
  item_id: string
  store_id: string
  branch_id: string
  name: string
  selection_type: 'single' | 'multiple'
  is_required: boolean
  min_selections: number
  max_selections: number
  modifiers: Modifier[]
}

export interface Modifier {
  id: string
  modifier_group_id: string
  store_id: string
  branch_id: string
  name: string
  price_adjustment: number
  is_available: boolean
  sort_order: number
}

export interface MenuItem extends Item {
  modifier_groups: ModifierGroup[]
}

export interface MenuCategory extends Category {
  items: MenuItem[]
}

export interface Menu {
  categories: MenuCategory[]
}

export interface BranchPreviewItem {
  id: string
  name: string
  image_url: string
  base_price: number
}

export interface DiscoverBranch {
  store_id: string
  store_name: string
  store_slug: string
  store_logo_url: string
  store_category: StoreCategory
  branch_id: string
  branch_name: string
  branch_address: string
  lat?: number | null
  lng?: number | null
  banner_image_url: string
  preview_items: BranchPreviewItem[]
}

export interface BranchDetail {
  store: Store
  branch: Branch
}

export interface Order {
  id: string
  order_number: number
  user_id: string
  store_id: string
  branch_id: string
  status: string
  total_price: number
  payment_method: string
  payment_status: string
  eta_minutes: number
  rejection_reason?: string
  created_at: string
  updated_at: string
  items: OrderItem[]
}

export interface OrderItem {
  id: string
  order_id: string
  item_name: string
  item_price: number
  quantity: number
  modifiers: OrderItemModifier[]
}

export interface OrderItemModifier {
  id: string
  modifier_name: string
  price_adjustment: number
}

export interface CartModifier {
  id: string
  name: string
  price: number
}

export interface CartItem {
  itemId: string
  imageUrl: string
  name: string
  price: number
  quantity: number
  modifiers: CartModifier[]
  totalPrice: number
}

export interface CartMeta {
  branchId: string
  branchName: string
  storeName: string
  bannerImageUrl: string
}

export interface DiscoverItem {
  id: string
  name: string
  description: string
  image_url: string
  base_price: number
  is_available: boolean
  created_at: string
  order_count: number
  has_required_modifiers: boolean
  branch_id: string
  branch_name: string
  branch_address: string
  lat?: number | null
  lng?: number | null
  store_id: string
  store_name: string
  store_category: StoreCategory
}

export interface FeedSection {
  title: string
  type: 'new' | 'popular'
  items: DiscoverItem[]
}

export interface FeedResponse {
  sections: FeedSection[]
}

export interface ItemsPageResponse {
  items: DiscoverItem[]
  total: number
  page: number
  limit: number
}

export interface BranchCart {
  branch: CartMeta
  items: CartItem[]
  paymentMethod: 'cash' | 'card'
  etaMinutes: number
}

export interface ItemDetailResponse {
  item: MenuItem
  branch: BranchDetail
}

export interface AuthUser {
  id: string
  telegram_id: number
  phone: string
  first_name: string
  last_name: string
  username: string
}
