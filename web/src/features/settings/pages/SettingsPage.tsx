import { Badge, Button, Group, Paper, SimpleGrid, Stack, Text, Title } from "@mantine/core";
import { IconArrowLeft, IconPlus, IconUser } from "@tabler/icons-react";
import { formatMoney, toTitleCase } from "../../../lib/formatters";
import type { User } from "../../../lib/api";
import type { Account, Profile } from "../../../types/domain";

interface SettingsPageProps {
  user: User;
  profile?: Profile;
  accounts: Account[];
  onBack: () => void;
  onCreateAccount: () => void;
}

export function SettingsPage({ user, profile, accounts, onBack, onCreateAccount }: SettingsPageProps) {
  return (
    <Stack gap="xl">
      <div>
        <Button variant="subtle" color="gray" size="compact-sm" px={0} mb="md" leftSection={<IconArrowLeft size={16} />} onClick={onBack}>
          Back to dashboard
        </Button>
        <Title order={1} className="page-title">Settings</Title>
        <Text c="dimmed" size="sm" mt={4}>Manage your profile and simulated accounts.</Text>
      </div>

      <Paper className="panel" p="xl">
        <Group gap="md">
          <IconUser size={24} />
          <div>
            <Text fw={650}>{profile ? `${profile.firstName} ${profile.lastName}` : "Fafnir investor"}</Text>
            <Text c="dimmed" size="sm">{user.email}</Text>
          </div>
        </Group>
      </Paper>

      <div>
        <Group justify="space-between" mb="md">
          <div><Text fw={650} fz="lg">Accounts</Text><Text c="dimmed" size="sm">Your paper money accounts</Text></div>
          <Button leftSection={<IconPlus size={16} />} onClick={onCreateAccount}>Open account</Button>
        </Group>
        <SimpleGrid cols={{ base: 1, sm: 2, xl: 3 }}>
          {accounts.map((account) => (
            <Paper className="panel" p="lg" key={account.id}>
              <Group justify="space-between"><Text fw={650}>{toTitleCase(account.type)}</Text><Badge variant="light" color="gray">{account.currency}</Badge></Group>
              <Text fz="xl" fw={650} mt="lg">{formatMoney(account.balance)}</Text>
              <Text c="dimmed" size="xs" mt={4}>•••• {account.accountNumber.slice(-4)}</Text>
            </Paper>
          ))}
        </SimpleGrid>
        {!accounts.length && <Paper className="panel" p="xl" mt="md"><Text fw={600}>No accounts yet</Text><Text c="dimmed" size="sm" mt={4}>Open one to start trading with simulated funds.</Text></Paper>}
      </div>
    </Stack>
  );
}
