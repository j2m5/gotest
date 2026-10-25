import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Login from './pages/auth/Login'
import Register from './pages/auth/Register'
import Home from './pages/auth/Home'
import Verify from './pages/auth/Verify'
import AccountProvider from './account/AccountProvider'
import AccountGate from './components/AccountGate'
import Characters from './pages/characters/Characters'
import NewCharacter from './pages/characters/NewCharacter'
import Game from './pages/characters/Game'

function App() {
  return (
    <AccountProvider><BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/verify" element={<Verify />} />
        <Route path="/characters" element={<AccountGate><Characters /></AccountGate>} />
        <Route path="/characters/new" element={<AccountGate><NewCharacter /></AccountGate>} />
        <Route path="/game" element={<AccountGate><Game /></AccountGate>} />
      </Routes>
    </BrowserRouter></AccountProvider>
  )
}

export default App
