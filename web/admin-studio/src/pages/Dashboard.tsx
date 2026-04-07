import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api, type Episode } from '../lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'

const STATUS_COLORS: Record<string, string> = {
  draft: 'text-gray-500',
  queued: 'text-blue-500',
  researching: 'text-yellow-500',
  brief_ready: 'text-orange-500',
  scripting: 'text-purple-500',
  voice_rendering: 'text-indigo-500',
  rendering: 'text-indigo-600',
  ready_for_review: 'text-teal-500',
  scheduled: 'text-blue-600',
  published: 'text-green-600',
  failed: 'text-red-600',
}

function StatusCount({ status, episodes }: { status: string; episodes: Episode[] }) {
  const count = episodes.filter((e) => e.status === status).length
  return (
    <div className="text-center">
      <div className={`text-3xl font-bold ${STATUS_COLORS[status] ?? 'text-gray-700'}`}>{count}</div>
      <div className="text-sm text-gray-500 capitalize mt-1">{status.replace(/_/g, ' ')}</div>
    </div>
  )
}

export default function Dashboard() {
  const { data: shows = [], isLoading: showsLoading } = useQuery({
    queryKey: ['shows'],
    queryFn: api.listShows,
  })

  const { data: episodes = [], isLoading: epsLoading } = useQuery({
    queryKey: ['episodes'],
    queryFn: () => api.listEpisodes(),
  })

  const statuses = ['draft', 'queued', 'researching', 'scripting', 'voice_rendering', 'rendering', 'published', 'failed']

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-500 mt-1">PodGen Network — overview</p>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="py-6">
            <div className="text-center">
              <div className="text-3xl font-bold text-indigo-600">
                {showsLoading ? '…' : shows.length}
              </div>
              <div className="text-sm text-gray-500 mt-1">Shows</div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-6">
            <div className="text-center">
              <div className="text-3xl font-bold text-indigo-600">
                {epsLoading ? '…' : episodes.length}
              </div>
              <div className="text-sm text-gray-500 mt-1">Total Episodes</div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-6">
            <div className="text-center">
              <div className="text-3xl font-bold text-green-600">
                {epsLoading ? '…' : episodes.filter((e) => e.status === 'published').length}
              </div>
              <div className="text-sm text-gray-500 mt-1">Published</div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="py-6">
            <div className="text-center">
              <div className="text-3xl font-bold text-red-600">
                {epsLoading ? '…' : episodes.filter((e) => e.status === 'failed').length}
              </div>
              <div className="text-sm text-gray-500 mt-1">Failed</div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Pipeline Status</CardTitle>
        </CardHeader>
        <CardContent>
          {epsLoading ? (
            <p className="text-gray-400">Loading…</p>
          ) : (
            <div className="grid grid-cols-4 md:grid-cols-8 gap-4 py-2">
              {statuses.map((s) => (
                <StatusCount key={s} status={s} episodes={episodes} />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <div className="grid md:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Recent Episodes</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {epsLoading ? (
              <p className="p-6 text-gray-400">Loading…</p>
            ) : episodes.length === 0 ? (
              <p className="p-6 text-gray-400">No episodes yet.</p>
            ) : (
              <ul className="divide-y divide-gray-100">
                {episodes.slice(0, 5).map((ep) => (
                  <li key={ep.id} className="px-6 py-3 flex items-center justify-between hover:bg-gray-50">
                    <Link to={`/episodes/${ep.id}`} className="text-sm font-medium text-indigo-600 hover:underline truncate max-w-xs">
                      {ep.title || ep.topic || 'Untitled'}
                    </Link>
                    <span className={`text-xs font-medium ${STATUS_COLORS[ep.status] ?? 'text-gray-500'}`}>
                      {ep.status.replace(/_/g, ' ')}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Shows</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {showsLoading ? (
              <p className="p-6 text-gray-400">Loading…</p>
            ) : shows.length === 0 ? (
              <p className="p-6 text-gray-400">No shows yet. <Link to="/shows" className="text-indigo-600 hover:underline">Create one</Link></p>
            ) : (
              <ul className="divide-y divide-gray-100">
                {shows.map((show) => (
                  <li key={show.id} className="px-6 py-3 flex items-center justify-between hover:bg-gray-50">
                    <Link to={`/shows/${show.id}`} className="text-sm font-medium text-indigo-600 hover:underline">
                      {show.name}
                    </Link>
                    <span className="text-xs text-gray-500">{show.niche}</span>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
