import { lazy, Suspense } from "react";
import { Center, Loader } from "@mantine/core";
import { Navigate, Route, Routes } from "react-router-dom";
import { NotFoundPage } from "./components/feedback/NotFoundPage";
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
      <Routes>
        {session.data ? (
          <>
            <Route path="/" element={<Dashboard user={session.data} />} />
            <Route path="/settings" element={<Dashboard user={session.data} />} />
            <Route path="/stocks/:symbol" element={<Dashboard user={session.data} />} />
          </>
        ) : (
          <>
            <Route path="/" element={<AuthPage />} />
            <Route path="/settings" element={<Navigate to="/" replace />} />
            <Route path="/stocks/:symbol" element={<Navigate to="/" replace />} />
          </>
        )}
        <Route path="*" element={<NotFoundPage authenticated={Boolean(session.data)} />} />
      </Routes>
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
