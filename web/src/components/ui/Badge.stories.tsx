import type { Meta, StoryObj } from '@storybook/react'
import { Badge } from './Badge'

const meta: Meta<typeof Badge> = {
  title: 'UI/Badge',
  component: Badge,
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'open', 'merged', 'closed', 'success', 'warning', 'agent', 'outline'],
    },
    pulse: { control: 'boolean' },
  },
}

export default meta
type Story = StoryObj<typeof Badge>

export const OpenPR: Story = {
  args: {
    children: 'Open',
    variant: 'open',
  },
}

export const MergedPR: Story = {
  args: {
    children: 'Merged',
    variant: 'merged',
  },
}

export const ClosedPR: Story = {
  args: {
    children: 'Closed',
    variant: 'closed',
  },
}

export const AgentProtocol: Story = {
  args: {
    children: 'Live Protocol',
    variant: 'agent',
    pulse: true,
  },
}

export const SystemSuccess: Story = {
  args: {
    children: 'System Operational',
    variant: 'success',
    pulse: true,
  },
}

export const WarningState: Story = {
  args: {
    children: 'Expiring Soon',
    variant: 'warning',
  },
}

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2 items-center">
      <Badge variant="default">Default</Badge>
      <Badge variant="open" pulse>Open</Badge>
      <Badge variant="merged">Merged</Badge>
      <Badge variant="closed">Closed</Badge>
      <Badge variant="success" pulse>Operational</Badge>
      <Badge variant="warning">Warning</Badge>
      <Badge variant="agent" pulse>MCP Agent</Badge>
      <Badge variant="outline">Outline</Badge>
    </div>
  ),
}
