import { useQuery } from "@tanstack/react-query";
import { SearchStocksDocument } from "../../../gql/graphql";
import { graphQLClient } from "../../../lib/api";

export function useStockSearch(query: string) {
  return useQuery({
    queryKey: ["stock-search", query],
    queryFn: () => graphQLClient.request(SearchStocksDocument, { query, limit: 8 }),
    enabled: query.length > 0,
    staleTime: 5 * 60_000,
    select: (data) => data.searchStocks,
  });
}
