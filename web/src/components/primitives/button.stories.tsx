import type { Meta, StoryObj } from '@storybook/react'
import { Button } from './button'
import { Terminal, ArrowRight } from 'lucide-react'

const meta: Meta<typeof Button> = {
  title: 'Primitives/Button',
  component: Button,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof Button>

export const Primary: Story = {
  args: {
    children: 'Deploy Capability Lease',
    variant: 'default',
  },
}

export const Secondary: Story = {
  args: {
    children: 'Inspect AST Rules',
    variant: 'secondary',
  },
}

export const Outline: Story = {
  args: {
    children: 'View Telemetry Ring',
    variant: 'outline',
  },
}

export const Destructive: Story = {
  args: {
    children: 'Revoke Lease',
    variant: 'destructive',
  },
}

export const WithIcon: Story = {
  args: {
    children: (
      <>
        <Terminal className="size-4 mr-1.5" />
        <span>Connect MCP</span>
        <ArrowRight className="size-4 ml-1.5" />
      </>
    ),
    variant: 'default',
  },
}

export const Disabled: Story = {
  args: {
    children: 'Direct Push Prohibited',
    isDisabled: true,
  },
}
