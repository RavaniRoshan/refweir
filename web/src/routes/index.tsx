import { createFileRoute } from '@tanstack/react-router';
import App from '@/App';

export interface IndexSearch {
  tab?: string;
  shell?: string;
}

export const Route = createFileRoute('/')({
  validateSearch: (search: Record<string, unknown>): IndexSearch => ({
    tab: typeof search.tab === 'string' ? search.tab : undefined,
    shell: typeof search.shell === 'string' ? search.shell : undefined,
  }),
  component: IndexPage,
});

function IndexPage() {
  return <App />;
}
