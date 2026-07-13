import {
  ActionIcon,
  AppShell,
  Avatar,
  Box,
  Burger,
  Group,
  Menu,
  Text,
  Tooltip,
  UnstyledButton,
} from "@mantine/core";
import {
  IconBell,
  IconChevronDown,
  IconLogout,
  IconRefresh,
  IconSettings,
} from "@tabler/icons-react";
import { useNavigate } from "react-router-dom";
import { Brand } from "../../../components/brand/Brand";
import type { User } from "../../../lib/api";
import type { Profile } from "../../../types/domain";
import { useLogout } from "../../auth/api/useLogout";
import { StockSearch } from "../../stocks/components/StockSearch";

interface DashboardHeaderProps {
  user: User;
  profile?: Profile;
  navOpened: boolean;
  refreshing: boolean;
  onToggleNav: () => void;
  onRefresh: () => void;
  onSelectStock: (symbol: string) => void;
}

export function DashboardHeader({
  user,
  profile,
  navOpened,
  refreshing,
  onToggleNav,
  onRefresh,
  onSelectStock,
}: DashboardHeaderProps) {
  const logout = useLogout();
  const navigate = useNavigate();
  const displayName = profile ? `${profile.firstName} ${profile.lastName}` : user.email;

  return (
    <AppShell.Header className="topbar">
      <Group h="100%" px={{ base: "md", md: "xl" }} justify="space-between" wrap="nowrap">
        <Group gap="md" wrap="nowrap">
          <Burger
            opened={navOpened}
            onClick={onToggleNav}
            hiddenFrom="md"
            size="sm"
          />
          <Brand />
          <StockSearch onSelect={onSelectStock} />
        </Group>

        <Group gap="xs" wrap="nowrap">
          <Tooltip label="Market data refreshes every 30 seconds">
            <ActionIcon
              variant="subtle"
              color="gray"
              onClick={onRefresh}
              loading={refreshing}
              aria-label="Refresh market data"
            >
              <IconRefresh size={19} />
            </ActionIcon>
          </Tooltip>
          <ActionIcon variant="subtle" color="gray" aria-label="Notifications">
            <IconBell size={19} />
          </ActionIcon>

          <Menu width={210} position="bottom-end">
            <Menu.Target>
              <UnstyledButton className="user-menu">
                <Group gap="sm" wrap="nowrap">
                  <Avatar color="lime" radius="xl" size={34}>
                    {getInitials(profile?.firstName, profile?.lastName, user.email)}
                  </Avatar>
                  <Box visibleFrom="sm">
                    <Text size="sm" fw={600} lh={1.2}>
                      {displayName}
                    </Text>
                    <Text size="xs" c="dimmed">
                      Paper account
                    </Text>
                  </Box>
                  <IconChevronDown size={14} />
                </Group>
              </UnstyledButton>
            </Menu.Target>
            <Menu.Dropdown>
              <Menu.Label>{user.email}</Menu.Label>
              <Menu.Item leftSection={<IconSettings size={16} />} onClick={() => navigate("/settings")}>Settings</Menu.Item>
              <Menu.Divider />
              <Menu.Item
                color="red"
                leftSection={<IconLogout size={16} />}
                onClick={() => logout.mutate()}
              >
                Sign out
              </Menu.Item>
            </Menu.Dropdown>
          </Menu>
        </Group>
      </Group>
    </AppShell.Header>
  );
}

function getInitials(first?: string, last?: string, email?: string) {
  return first
    ? `${first[0]}${last?.[0] ?? ""}`.toUpperCase()
    : (email?.slice(0, 2).toUpperCase() ?? "FA");
}
