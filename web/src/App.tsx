import { lazy, Suspense } from "react";
import { Center, Loader } from "@mantine/core";
import { useSession } from "./features/auth/api/useSession";

const AuthPage = lazy(() =>
  import("./features/auth/AuthPage").then((module) => ({ default: module.AuthPage })),
);
const Dashboard = lazy(() =>
  import("./features/dashboard/Dashboard").then((module) => ({ default: module.Dashboard })),
);

export function App() {
  const session = useSession();

  if (session.isPending) return <AppLoader />;

  return (
    <Suspense fallback={<AppLoader />}>
      {session.data ? <Dashboard user={session.data} /> : <AuthPage />}
    </Suspense>
  );
}

function AppLoader() {
  return (
    <Center h="100dvh" className="app-loading">
      <Loader size="sm" color="lime" />
    </Center>
  );
}
