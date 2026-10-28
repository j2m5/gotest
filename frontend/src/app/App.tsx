import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Login from '../features/auth/pages/Login'
import Register from '../features/auth/pages/Register'
import Home from './pages/Home'
import NotFound from './pages/NotFound'
import Verify from '../features/auth/pages/Verify'
import ForgotPassword from '../features/auth/pages/ForgotPassword'
import ResetPassword from '../features/auth/pages/ResetPassword'
import AccountProvider from './providers/AccountProvider'
import AccountGate from '../entities/account/AccountGate'
import Characters from '../features/character/pages/Characters'
import NewCharacter from '../features/character/pages/NewCharacter'
import Game from '../features/character/pages/Game'
import Profile from '../features/character/pages/Profile'
import Chat from '../features/social/pages/Chat'
import Moderation from '../features/social/pages/Moderation'
import Moderators from '../features/social/pages/Moderators'

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
        <Route path="/profile" element={<AccountGate><Profile /></AccountGate>} />
        <Route path="/chat" element={<AccountGate><Chat /></AccountGate>} />
        <Route path="/chat/moderators" element={<AccountGate><Moderators /></AccountGate>} />
        <Route path="/moderation" element={<AccountGate><Moderation /></AccountGate>} />
        <Route path="*" element={<NotFound />} />
      </Routes>
    </BrowserRouter></AccountProvider>
  )
}

export default App
