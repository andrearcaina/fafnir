import { Button, Center, Container, Stack, Text, Title } from "@mantine/core";
import { IconArrowLeft } from "@tabler/icons-react";
import { Link } from "react-router-dom";
import { Brand } from "../brand/Brand";

interface NotFoundPageProps {
  authenticated: boolean;
}

export function NotFoundPage({ authenticated }: NotFoundPageProps) {
  return (
    <Center mih="100dvh" px="md">
      <Container size="xs" w="100%">
        <Stack align="center" gap="md" ta="center">
          <Brand />
          <Text c="lime.4" fw={700} fz="sm" lts="0.12em">
            404
          </Text>
          <Title order={1}>That page does not exist.</Title>
          <Text c="dimmed" maw={440}>
            The address may be incorrect, or the page may have moved.
          </Text>
          <Button component={Link} to="/" replace leftSection={<IconArrowLeft size={17} />} mt="sm">
            {authenticated ? "Back to dashboard" : "Back to sign in"}
          </Button>
        </Stack>
      </Container>
    </Center>
  );
}
