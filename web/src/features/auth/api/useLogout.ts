import { notifications } from "@mantine/notifications";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { auth } from "../../../lib/api";
import { sessionQueryKey } from "./useSession";

export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: auth.logout,
    onSuccess: () => queryClient.setQueryData(sessionQueryKey, null),
    onError: (error) =>
      notifications.show({
        color: "red",
        title: "Could not sign out",
        message: error.message,
      }),
  });
}
