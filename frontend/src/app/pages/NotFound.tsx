import { Link } from 'react-router-dom'
import Layout from '../../shared/ui/Layout'

export default function NotFound() {
  return <Layout><h1>404</h1><p>Страница не найдена.</p><Link to="/">На главную</Link></Layout>
}
