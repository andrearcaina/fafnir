import { useMemo, useState } from "react";
import { Combobox, Group, Loader, Text, TextInput, ThemeIcon, useCombobox } from "@mantine/core";
import { IconChartLine, IconSearch } from "@tabler/icons-react";
import { useSupportedStocks } from "../api/useSupportedStocks";

interface StockSearchProps {
  onSelect: (symbol: string) => void;
}

export function StockSearch({ onSelect }: StockSearchProps) {
  const [search, setSearch] = useState("");
  const stocks = useSupportedStocks();
  const combobox = useCombobox({
    onDropdownClose: () => combobox.resetSelectedOption(),
  });

  const results = useMemo(() => {
    const query = search.trim().toUpperCase();
    const symbols = stocks.data ?? [];
    return query ? symbols.filter((symbol) => symbol.includes(query)) : symbols;
  }, [search, stocks.data]);

  const selectStock = (symbol: string) => {
    setSearch(symbol);
    combobox.closeDropdown();
    onSelect(symbol);
  };

  return (
    <Combobox store={combobox} onOptionSubmit={selectStock} withinPortal>
      <Combobox.Target>
        <TextInput
          visibleFrom="sm"
          className="market-search"
          value={search}
          leftSection={<IconSearch size={16} />}
          rightSection={stocks.isPending ? <Loader size={14} /> : undefined}
          placeholder="Search stocks"
          aria-label="Search supported stocks"
          onFocus={() => combobox.openDropdown()}
          onClick={() => combobox.openDropdown()}
          onChange={(event) => {
            setSearch(event.currentTarget.value);
            combobox.openDropdown();
            combobox.updateSelectedOptionIndex();
          }}
        />
      </Combobox.Target>

      <Combobox.Dropdown className="stock-search-dropdown">
        <Combobox.Header>
          <Group justify="space-between">
            <Text size="xs" fw={650}>Supported stocks</Text>
            <Text size="xs" c="dimmed">{results.length}</Text>
          </Group>
        </Combobox.Header>
        <Combobox.Options mah={360} style={{ overflowY: "auto" }}>
          {stocks.isError ? (
            <Combobox.Empty>Could not load supported stocks</Combobox.Empty>
          ) : results.length ? (
            results.map((symbol) => (
              <Combobox.Option value={symbol} key={symbol}>
                <Group gap="sm">
                  <ThemeIcon variant="light" color="gray" radius="xl" size="md">
                    <IconChartLine size={15} />
                  </ThemeIcon>
                  <div>
                    <Text size="sm" fw={650}>{symbol}</Text>
                    <Text size="xs" c="dimmed">Available to trade</Text>
                  </div>
                </Group>
              </Combobox.Option>
            ))
          ) : (
            <Combobox.Empty>No matching stocks</Combobox.Empty>
          )}
        </Combobox.Options>
      </Combobox.Dropdown>
    </Combobox>
  );
}
