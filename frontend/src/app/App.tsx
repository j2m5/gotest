import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Login from '../features/auth/pages/Login'
import Register from '../features/auth/pages/Register'
import Home from './pages/Home'
import Verify from '../features/auth/pages/Verify'
import ForgotPassword from '../features/auth/pages/ForgotPassword'
import ResetPassword from '../features/auth/pages/ResetPassword'
import AccountProvider from './providers/AccountProvider'
import AccountGate from '../entities/account/AccountGate'
import Characters from '../features/character/pages/Characters'
import NewCharacter from '../features/character/pages/NewCharacter'
import Game from '../features/character/pages/Game'

function App() {
  return (
    <AccountProvider><BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/verify" element={<Verify />} />
        <Route path="/forgot-password" element={<ForgotPassword />} />
        <Route path="/reset-password" element={<ResetPassword />} />
        <Route path="/characters" element={<AccountGate><Characters /></AccountGate>} />
        <Route path="/characters/new" element={<AccountGate><NewCharacter /></AccountGate>} />
        <Route path="/game" element={<AccountGate><Game /></AccountGate>} />
      </Routes>
    </BrowserRouter></AccountProvider>
  )
}

export default App
