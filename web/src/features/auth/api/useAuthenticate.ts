import { notifications } from "@mantine/notifications";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { auth } from "../../../lib/api";
import type { AuthFormValues, AuthMode } from "../types";
import { sessionQueryKey } from "./useSession";

export function useAuthenticate(mode: AuthMode) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (values: AuthFormValues) => {
      if (mode === "register") {
        await auth.register(values);
      }
      return auth.login({ email: values.email, password: values.password });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: sessionQueryKey });
      notifications.show({
        color: "lime",
        title: "Welcome to Fafnir",
        message: "Your paper portfolio is ready.",
      });
    },
    onError: (error) =>
      notifications.show({
        color: "red",
        title: "Could not sign in",
        message: error.message,
      }),
  });
}
