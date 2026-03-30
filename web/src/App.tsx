import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import CartPage from './pages/CartPage'
import HomePage from './pages/HomePage'
import ItemPage from './pages/ItemPage'
import OrderPage from './pages/OrderPage'
import OrdersPage from './pages/OrdersPage'
import BranchPage from './pages/BranchPage'
import StorePage from './pages/StorePage'

function App() {
  return (
    <BrowserRouter>
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
    </BrowserRouter>
  )
}

export default App
