import { Group, Stack, Text, ThemeIcon, Title } from "@mantine/core";
import { IconChartCandle, IconShieldLock } from "@tabler/icons-react";

export function AuthHero() {
  return (
    <Stack gap="xl" justify="center" className="auth-intro">
      <Text className="eyebrow">MARKET SIMULATOR</Text>
      <Title order={1} className="auth-title">
        Learn the market.
        <br />
        Risk nothing.
      </Title>
      <Text c="dimmed" maw={530} fz="lg" lh={1.65}>
        A focused place to track companies, test ideas, and understand every move in your
        portfolio.
      </Text>
      <Group gap="xl" mt="md">
        <Feature icon={<IconChartCandle size={19} />} label="Live market data" />
        <Feature icon={<IconShieldLock size={19} />} label="Private by default" />
      </Group>
    </Stack>
  );
}

function Feature({ icon, label }: { icon: React.ReactNode; label: string }) {
  return (
    <Group gap="sm">
      <ThemeIcon variant="light" color="gray" radius="xl">
        {icon}
      </ThemeIcon>
      <Text size="sm" fw={500}>
        {label}
      </Text>
    </Group>
  );
}
