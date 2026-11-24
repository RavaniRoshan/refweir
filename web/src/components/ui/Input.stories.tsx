import type { Meta, StoryObj } from '@storybook/react'
import { Input, SearchPill } from './Input'
import { Mail } from 'lucide-react'

const meta: Meta<typeof Input> = {
  title: 'UI/Input',
  component: Input,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof Input>

export const Default: Story = {
  args: {
    placeholder: 'Enter repository name...',
  },
}

export const WithIcon: Story = {
  args: {
    placeholder: 'developer@enterprise.internal',
    icon: <Mail className="w-3.5 h-3.5" />,
    type: 'email',
  },
}

export const WithError: Story = {
  args: {
    value: 'invalid#name!',
    error: 'Repository name can only contain alphanumeric characters and hyphens.',
  },
}

export const Disabled: Story = {
  args: {
    value: 'immutable-main-branch',
    disabled: true,
  },
}

export const SearchField: Story = {
  render: () => (
    <div className="w-64">
      <SearchPill />
    </div>
  ),
}
