import { BrowserRouter, Routes, Route, NavLink } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Dashboard from './pages/Dashboard'
import Shows from './pages/Shows'
import Episodes from './pages/Episodes'
import EpisodeDetail from './pages/EpisodeDetail'

const qc = new QueryClient({
  defaultOptions: { queries: { staleTime: 30_000 } },
})

const navClass = ({ isActive }: { isActive: boolean }) =>
  `block px-4 py-2 rounded-md text-sm font-medium transition-colors ${
    isActive
      ? 'bg-indigo-50 text-indigo-700'
      : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
  }`

function Layout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-gray-50 flex">
      <aside className="w-56 bg-white border-r border-gray-200 flex flex-col shrink-0">
        <div className="px-6 py-5 border-b border-gray-200">
          <h1 className="text-lg font-bold text-indigo-600">PodGen</h1>
          <p className="text-xs text-gray-400 mt-0.5">Admin Studio</p>
        </div>
        <nav className="flex-1 p-3 space-y-1">
          <NavLink to="/" end className={navClass}>Dashboard</NavLink>
          <NavLink to="/shows" className={navClass}>Shows</NavLink>
          <NavLink to="/episodes" className={navClass}>Episodes</NavLink>
        </nav>
        <div className="p-4 border-t border-gray-200">
          <p className="text-xs text-gray-400">PodGen Network MVP</p>
        </div>
      </aside>
      <main className="flex-1 p-8 overflow-auto">
        <div className="max-w-6xl mx-auto">{children}</div>
      </main>
    </div>
  )
}

export default function App() {
  return (
    <QueryClientProvider client={qc}>
      <BrowserRouter>
        <Layout>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/shows" element={<Shows />} />
            <Route path="/episodes" element={<Episodes />} />
            <Route path="/episodes/:episodeID" element={<EpisodeDetail />} />
          </Routes>
        </Layout>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
