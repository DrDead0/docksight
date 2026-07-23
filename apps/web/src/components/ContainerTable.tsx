import { StatusBadge } from '@/components/StatusBadge'
import type { Container } from '@/types/api'

type ContainerTableProps = {
  containers: Container[]
}

export function ContainerTable({ containers }: ContainerTableProps) {
  if (containers.length === 0) {
    return (
      <div className="rounded-lg border border-dashed border-border px-4 py-10 text-center text-sm text-muted-foreground">
        No containers discovered for this host yet.
      </div>
    )
  }

  return (
    <div className="overflow-x-auto rounded-lg border border-border">
      <table className="w-full min-w-[36rem] border-collapse text-left text-sm">
        <thead className="bg-muted/40 text-xs uppercase tracking-wide text-muted-foreground">
          <tr>
            <th className="px-4 py-3 font-medium">Name</th>
            <th className="px-4 py-3 font-medium">Image</th>
            <th className="px-4 py-3 font-medium">Status</th>
          </tr>
        </thead>
        <tbody>
          {containers.map((container) => (
            <tr
              key={container.id}
              className="border-t border-border/80 hover:bg-accent/30"
            >
              <td className="px-4 py-3 font-medium">{container.name}</td>
              <td className="px-4 py-3 text-muted-foreground">
                {container.image}
              </td>
              <td className="px-4 py-3">
                <StatusBadge status={container.state || container.status} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
