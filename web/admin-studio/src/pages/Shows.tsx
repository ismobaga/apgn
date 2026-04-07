import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api, type Show } from '../lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Badge } from '../components/ui/Badge'
import { formatDate } from '../lib/utils'

function CreateShowModal({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient()
  const [form, setForm] = useState({
    name: '',
    description: '',
    language: 'en',
    niche: '',
    tone: 'informative',
    cadence: 'weekly',
    default_duration_minutes: 30,
    default_format: 'solo',
  })

  const mutation = useMutation({
    mutationFn: () => api.createShow(form),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['shows'] })
      onClose()
    },
  })

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-lg p-6 space-y-4">
        <h2 className="text-xl font-semibold text-gray-900">Create Show</h2>
        <div className="grid grid-cols-2 gap-4">
          {(['name', 'niche', 'tone', 'cadence'] as const).map((field) => (
            <div key={field} className={field === 'name' ? 'col-span-2' : ''}>
              <label className="block text-sm font-medium text-gray-700 mb-1 capitalize">{field}</label>
              <input
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                value={String(form[field as keyof typeof form])}
                onChange={(e) => setForm((f) => ({ ...f, [field]: e.target.value }))}
              />
            </div>
          ))}
          <div className="col-span-2">
            <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              rows={3}
              value={form.description}
              onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
            />
          </div>
        </div>
        {mutation.error && (
          <p className="text-red-600 text-sm">{(mutation.error as Error).message}</p>
        )}
        <div className="flex justify-end gap-3">
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
          <Button onClick={() => mutation.mutate()} disabled={mutation.isPending || !form.name}>
            {mutation.isPending ? 'Creating…' : 'Create Show'}
          </Button>
        </div>
      </div>
    </div>
  )
}

export default function Shows() {
  const [showCreate, setShowCreate] = useState(false)
  const { data: shows = [], isLoading } = useQuery({
    queryKey: ['shows'],
    queryFn: api.listShows,
  })

  const statusVariant = (s: string) =>
    s === 'active' ? 'success' : s === 'paused' ? 'warning' : 'default'

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Shows</h1>
          <p className="text-gray-500 mt-1">{shows.length} show{shows.length !== 1 ? 's' : ''}</p>
        </div>
        <Button onClick={() => setShowCreate(true)}>+ New Show</Button>
      </div>

      {isLoading ? (
        <p className="text-gray-400">Loading…</p>
      ) : shows.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <p className="text-gray-500">No shows yet. Create your first show to get started.</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
          {shows.map((show: Show) => (
            <Card key={show.id} className="hover:shadow-md transition-shadow">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <CardTitle>{show.name}</CardTitle>
                  <Badge variant={statusVariant(show.status)}>{show.status}</Badge>
                </div>
                <p className="text-sm text-gray-500 mt-1">{show.slug}</p>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-gray-600 line-clamp-2">{show.description || 'No description.'}</p>
                <div className="mt-3 flex flex-wrap gap-2">
                  {show.niche && <Badge>{show.niche}</Badge>}
                  {show.tone && <Badge variant="info">{show.tone}</Badge>}
                  {show.cadence && <Badge variant="default">{show.cadence}</Badge>}
                </div>
                <div className="mt-4 flex items-center justify-between">
                  <span className="text-xs text-gray-400">{formatDate(show.created_at)}</span>
                  <Link to={`/shows/${show.id}`}>
                    <Button variant="ghost" size="sm">View →</Button>
                  </Link>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {showCreate && <CreateShowModal onClose={() => setShowCreate(false)} />}
    </div>
  )
}
