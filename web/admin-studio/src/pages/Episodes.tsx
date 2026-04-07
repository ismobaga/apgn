import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api, type Episode, type Show } from '../lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Badge } from '../components/ui/Badge'
import { formatDate } from '../lib/utils'

const STATUS_VARIANT: Record<string, 'default' | 'success' | 'warning' | 'error' | 'info'> = {
  draft: 'default',
  queued: 'info',
  researching: 'info',
  brief_ready: 'info',
  scripting: 'info',
  voice_rendering: 'info',
  rendering: 'info',
  ready_for_review: 'warning',
  scheduled: 'warning',
  published: 'success',
  failed: 'error',
}

function CreateEpisodeModal({ shows, onClose }: { shows: Show[]; onClose: () => void }) {
  const qc = useQueryClient()
  const [form, setForm] = useState({
    show_id: shows[0]?.id ?? '',
    topic: '',
    angle: '',
    target_duration_seconds: 1800,
  })

  const mutation = useMutation({
    mutationFn: () => api.createEpisode(form.show_id, form),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['episodes'] })
      onClose()
    },
  })

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-lg p-6 space-y-4">
        <h2 className="text-xl font-semibold text-gray-900">New Episode</h2>
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Show</label>
            <select
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              value={form.show_id}
              onChange={(e) => setForm((f) => ({ ...f, show_id: e.target.value }))}
            >
              {shows.map((s) => (
                <option key={s.id} value={s.id}>{s.name}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Topic</label>
            <input
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              value={form.topic}
              onChange={(e) => setForm((f) => ({ ...f, topic: e.target.value }))}
              placeholder="What is this episode about?"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Angle</label>
            <textarea
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              rows={2}
              value={form.angle}
              onChange={(e) => setForm((f) => ({ ...f, angle: e.target.value }))}
              placeholder="Unique perspective or angle"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Target Duration (minutes)</label>
            <input
              type="number"
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              value={Math.round(form.target_duration_seconds / 60)}
              onChange={(e) =>
                setForm((f) => ({ ...f, target_duration_seconds: parseInt(e.target.value) * 60 }))
              }
              min={5}
              max={120}
            />
          </div>
        </div>
        {mutation.error && (
          <p className="text-red-600 text-sm">{(mutation.error as Error).message}</p>
        )}
        <div className="flex justify-end gap-3">
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={() => mutation.mutate()} disabled={mutation.isPending || !form.topic || !form.show_id}>
            {mutation.isPending ? 'Creating…' : 'Create Episode'}
          </Button>
        </div>
      </div>
    </div>
  )
}

export default function Episodes() {
  const [showCreate, setShowCreate] = useState(false)
  const [filterStatus, setFilterStatus] = useState('')

  const { data: shows = [] } = useQuery({ queryKey: ['shows'], queryFn: api.listShows })
  const { data: episodes = [], isLoading } = useQuery({
    queryKey: ['episodes', filterStatus],
    queryFn: () => api.listEpisodes(filterStatus ? { status: filterStatus } : undefined),
  })

  const statuses = ['', 'draft', 'queued', 'researching', 'scripting', 'voice_rendering', 'rendering', 'ready_for_review', 'published', 'failed']

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Episodes</h1>
          <p className="text-gray-500 mt-1">{episodes.length} episode{episodes.length !== 1 ? 's' : ''}</p>
        </div>
        <Button onClick={() => setShowCreate(true)} disabled={shows.length === 0}>
          + New Episode
        </Button>
      </div>

      <div className="flex gap-2 flex-wrap">
        {statuses.map((s) => (
          <button
            key={s}
            onClick={() => setFilterStatus(s)}
            className={`px-3 py-1 rounded-full text-xs font-medium transition-colors ${
              filterStatus === s
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
          >
            {s || 'All'}
          </button>
        ))}
      </div>

      {isLoading ? (
        <p className="text-gray-400">Loading…</p>
      ) : episodes.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <p className="text-gray-500">No episodes found.</p>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>All Episodes</CardTitle>
          </CardHeader>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Episode</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Show</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                  <th className="px-6 py-3"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {episodes.map((ep: Episode) => {
                  const show = shows.find((s) => s.id === ep.show_id)
                  return (
                    <tr key={ep.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4">
                        <div className="font-medium text-gray-900">{ep.title || ep.topic || 'Untitled'}</div>
                        {ep.angle && <div className="text-xs text-gray-400 mt-0.5 truncate max-w-xs">{ep.angle}</div>}
                      </td>
                      <td className="px-6 py-4">
                        <Badge variant={STATUS_VARIANT[ep.status] ?? 'default'}>
                          {ep.status.replace(/_/g, ' ')}
                        </Badge>
                      </td>
                      <td className="px-6 py-4 text-gray-500">{show?.name ?? '—'}</td>
                      <td className="px-6 py-4 text-gray-400">{formatDate(ep.created_at)}</td>
                      <td className="px-6 py-4 text-right">
                        <Link to={`/episodes/${ep.id}`}>
                          <Button variant="ghost" size="sm">View →</Button>
                        </Link>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {showCreate && <CreateEpisodeModal shows={shows} onClose={() => setShowCreate(false)} />}
    </div>
  )
}
