import { Box, Group, Text } from "@mantine/core";

export function Brand() {
  return (
    <Group gap="sm" wrap="nowrap">
      <Box className="brand-mark">F</Box>
      <Text fw={700} fz="lg" lts="-0.02em">
        fafnir
      </Text>
    </Group>
  );
}
