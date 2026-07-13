import { Button, Paper, Stack, Text } from "@mantine/core";

interface ErrorPanelProps {
  message: string;
  onRetry: () => void;
}

export function ErrorPanel({ message, onRetry }: ErrorPanelProps) {
  return (
    <Paper className="panel" p={40}>
      <Stack align="center">
        <Text fw={650}>Could not load your dashboard</Text>
        <Text c="dimmed" size="sm" ta="center">
          {message}
        </Text>
        <Button variant="light" onClick={onRetry}>
          Try again
        </Button>
      </Stack>
    </Paper>
  );
}
