import { notifications } from "@mantine/notifications";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { auth } from "../../../lib/api";
import { sessionQueryKey } from "./useSession";

export function useDeleteUser() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  return useMutation({
    mutationFn: auth.deleteAccount,
    onSuccess: () => {
      queryClient.setQueryData(sessionQueryKey, null);
      queryClient.removeQueries({
        predicate: (query) => query.queryKey[0] !== sessionQueryKey[0],
      });
      navigate("/", { replace: true });
      notifications.show({ color: "lime", title: "Account deleted", message: "Your Fafnir account was removed." });
    },
    onError: (error) =>
      notifications.show({ color: "red", title: "Could not delete your account", message: error.message }),
  });
}
