import {
  Anchor,
  Button,
  Paper,
  PasswordInput,
  SimpleGrid,
  Stack,
  Text,
  TextInput,
  Title,
} from "@mantine/core";
import { useForm } from "@mantine/form";
import { IconArrowUpRight } from "@tabler/icons-react";
import { useAuthenticate } from "../api/useAuthenticate";
import type { AuthFormValues, AuthMode } from "../types";

interface AuthFormProps {
  mode: AuthMode;
  onModeChange: (mode: AuthMode) => void;
}

const initialValues: AuthFormValues = {
  firstName: "",
  lastName: "",
  email: "",
  password: "",
};

export function AuthForm({ mode, onModeChange }: AuthFormProps) {
  const authenticate = useAuthenticate(mode);
  const form = useForm({
    mode: "uncontrolled",
    initialValues,
    validate: {
      firstName: (value) =>
        mode === "register" && !value.trim() ? "First name is required" : null,
      lastName: (value) =>
        mode === "register" && !value.trim() ? "Last name is required" : null,
      email: (value) => (/^\S+@\S+\.\S+$/.test(value) ? null : "Enter a valid email"),
      password: (value) => (value.length >= 6 ? null : "Use at least 6 characters"),
    },
  });

  const toggleMode = () => {
    onModeChange(mode === "login" ? "register" : "login");
    form.clearErrors();
  };

  return (
    <Paper className="auth-card" radius="lg" p={{ base: "xl", sm: 38 }}>
      <Stack gap="lg">
        <div>
          <Title order={2}>{mode === "login" ? "Welcome back" : "Start investing"}</Title>
          <Text c="dimmed" mt={6}>
            {mode === "login"
              ? "Sign in to view your portfolio."
              : "Create your paper trading account."}
          </Text>
        </div>

        <form onSubmit={form.onSubmit((values) => authenticate.mutate(values))}>
          <Stack gap="md">
            {mode === "register" && (
              <SimpleGrid cols={2} spacing="sm">
                <TextInput
                  label="First name"
                  placeholder="Sigurd"
                  key={form.key("firstName")}
                  {...form.getInputProps("firstName")}
                />
                <TextInput
                  label="Last name"
                  placeholder="Volsung"
                  key={form.key("lastName")}
                  {...form.getInputProps("lastName")}
                />
              </SimpleGrid>
            )}
            <TextInput
              label="Email"
              placeholder="you@example.com"
              key={form.key("email")}
              {...form.getInputProps("email")}
            />
            <PasswordInput
              label="Password"
              placeholder="At least 6 characters"
              key={form.key("password")}
              {...form.getInputProps("password")}
            />
            <Button
              type="submit"
              size="md"
              mt="sm"
              loading={authenticate.isPending}
              rightSection={<IconArrowUpRight size={18} />}
            >
              {mode === "login" ? "Sign in" : "Create account"}
            </Button>
          </Stack>
        </form>

        <Text ta="center" c="dimmed" size="sm">
          {mode === "login" ? "New to Fafnir? " : "Already have an account? "}
          <Anchor component="button" type="button" c="lime" fw={600} onClick={toggleMode}>
            {mode === "login" ? "Create an account" : "Sign in"}
          </Anchor>
        </Text>
      </Stack>
    </Paper>
  );
}
