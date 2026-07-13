import { useState } from "react";
import { Box, Container, Group, SimpleGrid, Text } from "@mantine/core";
import { Brand } from "../../components/brand/Brand";
import { AuthForm } from "./components/AuthForm";
import { AuthHero } from "./components/AuthHero";
import type { AuthMode } from "./types";
import "./auth.css";

export function AuthPage() {
  const [mode, setMode] = useState<AuthMode>("login");

  return (
    <Box className="auth-shell">
      <Container size="xl" h="100%">
        <Group justify="space-between" py="xl">
          <Brand />
          <Text c="dimmed" size="sm" visibleFrom="sm">
            Paper trading, real market instincts.
          </Text>
        </Group>

        <SimpleGrid
          cols={{ base: 1, md: 2 }}
          spacing={{ base: 48, md: 100 }}
          className="auth-grid"
        >
          <AuthHero />
          <AuthForm mode={mode} onModeChange={setMode} />
        </SimpleGrid>
      </Container>
    </Box>
  );
}
