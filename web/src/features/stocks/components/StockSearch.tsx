import { useState } from "react";
import { Combobox, Group, Loader, Text, TextInput, ThemeIcon, useCombobox } from "@mantine/core";
import { useDebouncedValue } from "@mantine/hooks";
import { IconChartLine, IconSearch } from "@tabler/icons-react";
import { useStockSearch } from "../api/useStockSearch";

interface StockSearchProps {
  onSelect: (symbol: string) => void;
}

export function StockSearch({ onSelect }: StockSearchProps) {
  const [search, setSearch] = useState("");
  const [debouncedSearch] = useDebouncedValue(search.trim(), 250);
  const stocks = useStockSearch(debouncedSearch);
  const combobox = useCombobox({
    onDropdownClose: () => combobox.resetSelectedOption(),
  });
  const results = stocks.data ?? [];

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
          rightSection={stocks.isFetching ? <Loader size={14} /> : undefined}
          placeholder="Search Yahoo Finance"
          aria-label="Search market symbols"
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
            <Text size="xs" fw={650}>Yahoo Finance</Text>
            <Text size="xs" c="dimmed">{results.length} results</Text>
          </Group>
        </Combobox.Header>
        <Combobox.Options mah={360} style={{ overflowY: "auto" }}>
          {!debouncedSearch ? (
            <Combobox.Empty>Type a symbol or company name</Combobox.Empty>
          ) : stocks.isError ? (
            <Combobox.Empty>Could not search Yahoo Finance</Combobox.Empty>
          ) : results.length ? (
            results.map((result) => (
              <Combobox.Option value={result.symbol} key={`${result.symbol}:${result.exchange}`}>
                <Group gap="sm" wrap="nowrap">
                  <ThemeIcon variant="light" color="gray" radius="xl" size="md">
                    <IconChartLine size={15} />
                  </ThemeIcon>
                  <div style={{ minWidth: 0 }}>
                    <Group gap="xs" wrap="nowrap">
                      <Text size="sm" fw={650}>{result.symbol}</Text>
                      <Text size="xs" c="dimmed">{result.instrumentType}</Text>
                    </Group>
                    <Text size="xs" c="dimmed" truncate>{result.name || result.exchangeFullName || result.exchange}</Text>
                  </div>
                </Group>
              </Combobox.Option>
            ))
          ) : stocks.isFetching ? (
            <Combobox.Empty>Searching…</Combobox.Empty>
          ) : (
            <Combobox.Empty>No matching market symbols</Combobox.Empty>
          )}
        </Combobox.Options>
      </Combobox.Dropdown>
    </Combobox>
  );
}
