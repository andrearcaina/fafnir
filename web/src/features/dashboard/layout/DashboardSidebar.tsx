import {
  AppShell,
  Badge,
  Divider,
  NavLink,
  Paper,
  Stack,
  Text,
  ThemeIcon,
} from "@mantine/core";
import {
  IconBook,
  IconBriefcase2,
  IconChartLine,
  IconStar,
  IconTrendingUp,
  IconWallet,
} from "@tabler/icons-react";
import { formatMoney } from "../../../lib/formatters";
import type { Account } from "../../../types/domain";
import type { DashboardSection } from "../types";

interface DashboardSidebarProps {
  section: DashboardSection;
  accounts: Account[];
  orderCount: number;
  onNavigate: (section: DashboardSection) => void;
}

const navigation = [
  { label: "Overview", icon: IconChartLine },
  { label: "Portfolio", icon: IconBriefcase2 },
  { label: "Orders", icon: IconBook },
  { label: "Watchlist", icon: IconStar },
] satisfies { label: DashboardSection; icon: typeof IconChartLine }[];

export function DashboardSidebar({
  section,
  accounts,
  orderCount,
  onNavigate,
}: DashboardSidebarProps) {
  return (
    <AppShell.Navbar className="sidebar" p="md">
      <AppShell.Section grow>
        <Stack gap={5}>
          {navigation.map(({ label, icon: Icon }) => (
            <NavLink
              key={label}
              label={label}
              leftSection={<Icon size={19} />}
              active={section === label}
              onClick={() => onNavigate(label)}
              rightSection={
                label === "Orders" && orderCount ? (
                  <Badge size="xs" variant="light" color="gray">
                    {orderCount}
                  </Badge>
                ) : undefined
              }
            />
          ))}
        </Stack>

        <Divider my="lg" />
        <Text className="nav-caption">ACCOUNTS</Text>
        <Stack gap={4} mt="sm">
          {accounts.map((account) => (
            <NavLink
              key={account.id}
              label={account.type.toLowerCase()}
              description={`${account.currency} ${formatMoney(account.balance)}`}
              leftSection={
                <ThemeIcon variant="light" color="gray" radius="md">
                  <IconWallet size={16} />
                </ThemeIcon>
              }
            />
          ))}
          {!accounts.length && (
            <Text size="xs" c="dimmed" px="sm">
              No funded accounts yet
            </Text>
          )}
        </Stack>
      </AppShell.Section>

      <AppShell.Section>
        <Paper className="learn-card" p="md" radius="md">
          <ThemeIcon color="lime" variant="light" radius="xl">
            <IconTrendingUp size={17} />
          </ThemeIcon>
          <Text fw={600} size="sm" mt="sm">
            Build your instincts
          </Text>
          <Text c="dimmed" size="xs" mt={4}>
            Every trade here is simulated.
          </Text>
        </Paper>
      </AppShell.Section>
    </AppShell.Navbar>
  );
}
