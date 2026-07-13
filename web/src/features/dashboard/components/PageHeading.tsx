import { Button, Group, Text, Title } from "@mantine/core";
import { IconArrowDown, IconPlus } from "@tabler/icons-react";
import type { DashboardSection } from "../types";

interface PageHeadingProps {
  section: DashboardSection;
  firstName?: string;
  onTrade: () => void;
  onDeposit: () => void;
  onCreateAccount: () => void;
  hasAccounts: boolean;
}

export function PageHeading({ section, firstName, onTrade, onDeposit, onCreateAccount, hasAccounts }: PageHeadingProps) {
  return (
    <Group justify="space-between" align="flex-end" mb="xl">
      <div>
        <Text c="dimmed" size="sm" mb={4}>
          {getGreeting()}, {firstName ?? "investor"}
        </Text>
        <Title order={1} className="page-title">
          {section}
        </Title>
      </div>
      <Group gap="sm">
        {hasAccounts ? (
          <Button variant="default" leftSection={<IconArrowDown size={16} />} onClick={onDeposit}>Deposit</Button>
        ) : (
          <Button variant="default" leftSection={<IconPlus size={16} />} onClick={onCreateAccount}>Open account</Button>
        )}
        <Button leftSection={<IconPlus size={17} />} onClick={onTrade}>Trade</Button>
      </Group>
    </Group>
  );
}

function getGreeting() {
  const hour = new Date().getHours();
  if (hour < 12) return "Good morning";
  if (hour < 18) return "Good afternoon";
  return "Good evening";
}
