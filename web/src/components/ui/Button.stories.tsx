import type { Meta, StoryObj } from '@storybook/react'
import { Button } from './Button'
import { GitPullRequest, Trash2, ArrowRight } from 'lucide-react'

const meta: Meta<typeof Button> = {
  title: 'UI/Button',
  component: Button,
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary', 'outline', 'ghost', 'destructive'],
    },
    size: {
      control: 'select',
      options: ['xs', 'sm', 'md', 'lg', 'icon'],
    },
    disabled: { control: 'boolean' },
    loading: { control: 'boolean' },
  },
}

export default meta
type Story = StoryObj<typeof Button>

export const Primary: Story = {
  args: {
    children: 'Primary Action',
    variant: 'primary',
    size: 'md',
  },
}

export const Secondary: Story = {
  args: {
    children: 'Secondary Action',
    variant: 'secondary',
    size: 'md',
  },
}

export const Outline: Story = {
  args: {
    children: 'Outline Button',
    variant: 'outline',
    size: 'md',
  },
}

export const Destructive: Story = {
  args: {
    children: 'Delete Repository',
    variant: 'destructive',
    size: 'md',
  },
}

export const LoadingState: Story = {
  args: {
    children: 'Deploying Contract...',
    variant: 'primary',
    loading: true,
  },
}

export const DisabledState: Story = {
  args: {
    children: 'Merge Blocked',
    variant: 'primary',
    disabled: true,
  },
}

export const WithIcons: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <Button variant="primary">
        <GitPullRequest className="w-3.5 h-3.5 mr-1" />
        <span>Create Pull Request</span>
      </Button>
      <Button variant="outline">
        <span>Proceed</span>
        <ArrowRight className="w-3.5 h-3.5 ml-1" />
      </Button>
      <Button variant="destructive" size="icon">
        <Trash2 className="w-3.5 h-3.5" />
      </Button>
    </div>
  ),
}

export const AllSizes: Story = {
  render: () => (
    <div className="flex items-center gap-2">
      <Button size="xs">Extra Small (xs)</Button>
      <Button size="sm">Small (sm)</Button>
      <Button size="md">Medium (md)</Button>
      <Button size="lg">Large (lg)</Button>
    </div>
  ),
}
