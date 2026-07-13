import { Text } from "@mantine/core";

interface PriceChangeProps {
  value: number;
  compact?: boolean;
}

export function PriceChange({ value, compact = false }: PriceChangeProps) {
  const positive = value >= 0;

  return (
    <Text
      component="span"
      c={positive ? "lime.5" : "red.5"}
      size={compact ? "xs" : "sm"}
      fw={600}
    >
      {positive ? "+" : ""}
      {value.toFixed(2)}%
    </Text>
  );
}
