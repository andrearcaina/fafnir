import { Stack, Text, ThemeIcon } from "@mantine/core";
import { IconShoppingBag } from "@tabler/icons-react";

interface EmptyStateProps {
  title: string;
  detail: string;
}

export function EmptyState({ title, detail }: EmptyStateProps) {
  return (
    <Stack align="center" justify="center" py={50} gap={6}>
      <ThemeIcon variant="light" color="gray" radius="xl" size="lg">
        <IconShoppingBag size={18} />
      </ThemeIcon>
      <Text fw={600} size="sm">
        {title}
      </Text>
      <Text c="dimmed" size="xs" ta="center">
        {detail}
      </Text>
    </Stack>
  );
}
