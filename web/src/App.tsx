import { BrowserRouter, Routes, Route } from 'react-router-dom'
import StorePage from './pages/StorePage'
import CartPage from './pages/CartPage'
import OrderPage from './pages/OrderPage'
import OrdersPage from './pages/OrdersPage'

function App() {
  return (
    <BrowserRouter>
      <div className="min-h-screen bg-gray-50">
        <Routes>
          <Route path="/:slug" element={<StorePage />} />
          <Route path="/cart" element={<CartPage />} />
          <Route path="/order/:id" element={<OrderPage />} />
          <Route path="/orders" element={<OrdersPage />} />
          <Route
            path="/"
            element={
              <div className="flex items-center justify-center min-h-screen">
                <div className="text-center">
                  <h1 className="text-3xl font-bold mb-2">Xpressgo</h1>
                  <p className="text-gray-500">Open this app from a Telegram bot</p>
                </div>
              </div>
            }
          />
        </Routes>
      </div>
    </BrowserRouter>
  )
}

export default App
