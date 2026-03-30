import { Suspense, lazy } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'

const CartPage = lazy(() => import('./pages/CartPage'))
const HomePage = lazy(() => import('./pages/HomePage'))
const ItemPage = lazy(() => import('./pages/ItemPage'))
const OrderPage = lazy(() => import('./pages/OrderPage'))
const OrdersPage = lazy(() => import('./pages/OrdersPage'))
const BranchPage = lazy(() => import('./pages/BranchPage'))
const StorePage = lazy(() => import('./pages/StorePage'))

function App() {
  return (
    <BrowserRouter>
      <Suspense fallback={<div className="flex min-h-screen items-center justify-center">Loading...</div>}>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/branch/:id" element={<BranchPage />} />
          <Route path="/item/:id" element={<ItemPage />} />
          <Route path="/cart" element={<CartPage />} />
          <Route path="/order/:id" element={<OrderPage />} />
          <Route path="/orders" element={<OrdersPage />} />
          <Route path="/:slug" element={<StorePage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  )
}

export default App
