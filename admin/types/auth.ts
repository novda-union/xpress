export type StaffRole = 'director' | 'manager' | 'barista'

export interface Staff {
  id: string
  store_id: string
  branch_id: string | null
  branch_name?: string | null
  staff_code: string
  name: string
  role: StaffRole
  is_active?: boolean
  created_at?: string
}

export interface StaffGroup {
  branch_id: string | null
  branch_name: string
  staff: Staff[]
}

export interface BranchSummary {
  id: string
  store_id: string
  name: string
  address: string
  lat: number | null
  lng: number | null
  banner_image_url: string
  telegram_group_chat_id: number | null
  is_active: boolean
  staff_count?: number
  created_at?: string
  updated_at?: string
}

export interface AdminOrderItemModifier {
  id: string
  modifier_name: string
  price_adjustment: number
}

export interface AdminOrderItem {
  id: string
  item_name: string
  item_price: number
  quantity: number
  modifiers: AdminOrderItemModifier[]
}

export interface AdminOrder {
  id: string
  order_number: number
  store_id: string
  branch_id: string
  status: string
  total_price: number
  payment_method: string
  eta_minutes: number
  created_at: string
  items: AdminOrderItem[]
}

export interface AuthState {
  token: string | null
  staff: Staff | null
  initialized: boolean
}
