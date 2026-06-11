import { SharedResourcePage } from '@/components/share/shared-resource-page';

export const metadata = {
  title: 'Shared AgentTrace view',
  description: 'Read-only redacted trace or replay view',
  robots: {
    index: false,
    follow: false,
  },
};

interface SharedPageProps {
  params: Promise<{ token: string }>;
}

export default async function SharePage({ params }: SharedPageProps) {
  const { token } = await params;
  return <SharedResourcePage token={token} />;
}
