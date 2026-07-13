import { useQuery } from "@tanstack/react-query";
import { auth } from "../../../lib/api";

export const sessionQueryKey = ["session"] as const;

export function useSession() {
  return useQuery({
    queryKey: sessionQueryKey,
    queryFn: auth.me,
    retry: false,
    staleTime: 5 * 60_000,
  });
}
