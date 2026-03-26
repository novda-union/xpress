export interface Store {
  id: string
  name: string
  slug: string
  description: string
  address: string
  phone: string
  logo_url: string
}

export interface Category {
  id: string
  store_id: string
  name: string
  sort_order: number
  is_active: boolean
}

export interface Item {
  id: string
  category_id: string
  store_id: string
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
  name: string
  price_adjustment: number
  is_available: boolean
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

export interface Order {
  id: string
  order_number: number
  user_id: string
  store_id: string
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

export interface CartItem {
  itemId: string
  name: string
  price: number
  quantity: number
  modifiers: { id: string; name: string; price: number }[]
  totalPrice: number
}
