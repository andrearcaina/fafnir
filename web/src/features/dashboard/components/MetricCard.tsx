import { Paper, Skeleton, Text } from "@mantine/core";

interface MetricCardProps {
  label: string;
  value: string;
  detail: string;
  loading: boolean;
  featured?: boolean;
}

export function MetricCard({
  label,
  value,
  detail,
  loading,
  featured = false,
}: MetricCardProps) {
  return (
    <Paper className="metric-card" data-featured={featured || undefined} p="lg">
      <Text c="dimmed" size="sm">
        {label}
      </Text>
      {loading ? (
        <Skeleton h={34} w="55%" my={8} />
      ) : (
        <Text className="metric-value">{value}</Text>
      )}
      <Text c="dimmed" size="xs">
        {detail}
      </Text>
    </Paper>
  );
}
