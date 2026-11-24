import type { Meta, StoryObj } from '@storybook/react'
import { Badge } from './badge'
import { CheckCircle2, AlertTriangle, ShieldCheck } from 'lucide-react'

const meta: Meta<typeof Badge> = {
  title: 'Primitives/Badge',
  component: Badge,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof Badge>

export const Default: Story = {
  args: {
    children: 'agent:refweir-worker',
    variant: 'default',
  },
}

export const Secondary: Story = {
  args: {
    children: 'SHA-256 Verified',
    variant: 'secondary',
  },
}

export const Outline: Story = {
  args: {
    children: 'refs/heads/agent/*',
    variant: 'outline',
  },
}

export const Destructive: Story = {
  args: {
    children: 'GH007 Direct Push Blocked',
    variant: 'destructive',
  },
}

export const WithIcon: Story = {
  render: () => (
    <div className="flex gap-2">
      <Badge variant="default" className="gap-1">
        <ShieldCheck className="size-3" />
        Policy Gate Passed
      </Badge>
      <Badge variant="secondary" className="gap-1">
        <CheckCircle2 className="size-3 text-emerald-500" />
        3-Way Merge Clean
      </Badge>
      <Badge variant="destructive" className="gap-1">
        <AlertTriangle className="size-3" />
        Secret Detected
      </Badge>
    </div>
  ),
}
