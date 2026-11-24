import type { Meta, StoryObj } from '@storybook/react'
import { Dialog } from './Dialog'
import { Button } from './Button'
import { ShieldCheck } from 'lucide-react'

const meta: Meta<typeof Dialog> = {
  title: 'UI/Dialog',
  component: Dialog,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof Dialog>

export const DefaultOpen: Story = {
  args: {
    isOpen: true,
    title: 'Ephemeral Capability Lease',
    description: 'Grant scoped capabilities to autonomous worker agent for 30 minutes.',
    children: (
      <div className="p-4 space-y-4">
        <div className="p-3 bg-secondary/30 rounded-lg text-xs space-y-1 border border-border">
          <div className="flex items-center gap-1.5 text-primary font-medium">
            <ShieldCheck className="w-4 h-4" />
            <span>Cryptographic Ref Isolation</span>
          </div>
          <p className="text-muted-foreground font-mono">Allowed: refs/heads/agent/*</p>
          <p className="text-muted-foreground font-mono">Forbidden: refs/heads/main (Blocked at transport)</p>
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <Button variant="outline" size="sm">Cancel</Button>
          <Button size="sm">Issue HMAC Lease</Button>
        </div>
      </div>
    ),
  },
}
