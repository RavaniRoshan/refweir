import { Component, type ErrorInfo, type ReactNode } from 'react'
import { AlertCircle, RotateCcw } from 'lucide-react'
import { Button } from './Button'

export interface ViewErrorBoundaryProps {
  children: ReactNode
  fallbackTitle?: string
}

interface ViewErrorBoundaryState {
  hasError: boolean
  error: Error | null
}

export class ViewErrorBoundary extends Component<ViewErrorBoundaryProps, ViewErrorBoundaryState> {
  constructor(props: ViewErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): ViewErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ViewErrorBoundary caught error:', error, errorInfo)
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null })
  }

  handleReload = () => {
    if (typeof window !== 'undefined') {
      window.location.reload()
    }
  }

  render() {
    if (this.state.hasError) {
      return (
        <div
          role="alert"
          aria-live="assertive"
          className="min-h-[400px] flex flex-col items-center justify-center p-8 border border-border bg-card rounded-md text-center max-w-xl mx-auto my-8 space-y-4"
        >
          <div className="w-10 h-10 rounded-full bg-destructive/10 text-destructive flex items-center justify-center">
            <AlertCircle className="w-5 h-5" aria-hidden="true" />
          </div>
          <div className="space-y-1">
            <h3 className="text-sm font-semibold text-foreground">
              {this.props.fallbackTitle || 'Failed to load view'}
            </h3>
            <p className="text-xs text-muted-foreground">
              {this.state.error?.message?.toLowerCase().includes('fetch') ||
               this.state.error?.message?.toLowerCase().includes('import') ||
               this.state.error?.message?.toLowerCase().includes('chunk')
                ? 'A new version of Refweir may have been deployed. Please reload the page to get the latest version.'
                : (this.state.error?.message || 'An unexpected error occurred while loading this view.')}
            </p>
          </div>
          <div className="flex items-center gap-2 pt-2">
            <Button variant="primary" size="sm" onClick={this.handleReload} className="gap-1.5">
              <RotateCcw className="w-3.5 h-3.5" />
              <span>Reload page</span>
            </Button>
            <Button variant="secondary" size="sm" onClick={this.handleReset}>
              <span>Try again</span>
            </Button>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
