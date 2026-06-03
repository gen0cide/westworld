// ItemSprite renders an item's authentic RSC inventory icon via GET /sprite.
// If the sprite 404s (item id outside the classic config85 range, or archives
// unavailable), it falls back to a short text stub so the cell stays readable.

import { useState } from 'react'

function shortName(n: string): string { return n.length > 8 ? n.slice(0, 7) + '…' : n }

export function ItemSprite({ id, name }: { id: number; name?: string }) {
  const [failed, setFailed] = useState(false)
  if (failed || !id) {
    return <span className="stub">{shortName(name ?? `#${id}`)}</span>
  }
  return (
    <img
      className="icon"
      src={`/sprite?kind=item&id=${id}`}
      alt={name ?? String(id)}
      draggable={false}
      onError={() => setFailed(true)}
    />
  )
}
