import { useState } from "react";
import { AppShell, Box, Drawer } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { useMatch, useNavigate } from "react-router-dom";
import { ErrorPanel } from "../../components/feedback/ErrorPanel";
import type { User } from "../../lib/api";
import { OrderTicket } from "../trading/components/OrderTicket";
import { StockDetailPage } from "../stocks/pages/StockDetailPage";
import { DepositDialog } from "../funding/components/DepositDialog";
import { CreateAccountDialog } from "../accounts/components/CreateAccountDialog";
import { SettingsPage } from "../settings/pages/SettingsPage";
import { useDashboardData } from "./api/useDashboardData";
import { PageHeading } from "./components/PageHeading";
import { DashboardHeader } from "./layout/DashboardHeader";
import { DashboardSidebar } from "./layout/DashboardSidebar";
import { DashboardContent } from "./sections/DashboardContent";
import type { ChartPeriod, DashboardSection } from "./types";
import "@mantine/charts/styles.css";
import "./dashboard.css";

const DEFAULT_SYMBOL = "AAPL";

export function Dashboard({ user }: { user: User }) {
  const [navOpened, nav] = useDisclosure(false);
  const [ticketOpened, ticket] = useDisclosure(false);
  const [section, setSection] = useState<DashboardSection>("Overview");
  const [period, setPeriod] = useState<ChartPeriod>("1M");
  const [depositOpened, deposit] = useDisclosure(false);
  const [createAccountOpened, createAccount] = useDisclosure(false);
  const navigateTo = useNavigate();
  const stockMatch = useMatch("/stocks/:symbol");
  const settingsMatch = useMatch("/settings");
  const activeSymbol = stockMatch?.params.symbol?.toUpperCase() ?? DEFAULT_SYMBOL;
  const data = useDashboardData(activeSymbol, period);

  const navigate = (nextSection: DashboardSection) => {
    setSection(nextSection);
    navigateTo("/");
    nav.close();
  };

  const openStock = (symbol: string) => {
    navigateTo(`/stocks/${symbol.toUpperCase()}`);
    nav.close();
  };

  const completeOrder = () => {
    ticket.close();
    void data.refresh();
  };

  return (
    <AppShell
      header={{ height: 68 }}
      navbar={{ width: 238, breakpoint: "md", collapsed: { mobile: !navOpened } }}
      padding={0}
      className="dashboard-shell"
    >
      <DashboardHeader
        user={user}
        profile={data.profile}
        navOpened={navOpened}
        refreshing={data.isRefreshing}
        onToggleNav={nav.toggle}
        onRefresh={() => void data.refresh()}
        onSelectStock={openStock}
      />
      <DashboardSidebar
        section={section}
        accounts={data.accounts}
        orderCount={data.orders.length}
        onNavigate={navigate}
      />

      <AppShell.Main>
        <Box className="dashboard-main">
          {data.error ? (
            <ErrorPanel message={data.error.message} onRetry={() => void data.refresh()} />
          ) : settingsMatch ? (
            <SettingsPage
              user={user}
              profile={data.profile}
              accounts={data.accounts}
              onBack={() => navigateTo("/")}
              onCreateAccount={createAccount.open}
            />
          ) : stockMatch ? (
            <StockDetailPage
              symbol={activeSymbol}
              quote={data.activeQuote}
              metadata={data.activeMetadata}
              chartData={data.chartData}
              chartLoading={data.isChartLoading}
              period={period}
              onPeriodChange={setPeriod}
              onBack={() => navigateTo("/")}
              onTrade={ticket.open}
              isWatchlisted={data.watchlistSymbols.includes(activeSymbol)}
            />
          ) : (
            <>
              <PageHeading
                section={section}
                firstName={data.profile?.firstName}
                onTrade={ticket.open}
                onDeposit={deposit.open}
                onCreateAccount={createAccount.open}
                hasAccounts={data.accounts.length > 0}
              />
              <DashboardContent
                section={section}
                data={data}
                activeSymbol={activeSymbol}
                period={period}
                onSymbolChange={openStock}
                onPeriodChange={setPeriod}
                onWatchlistSelect={openStock}
                onTrade={ticket.open}
              />
            </>
          )}
        </Box>
      </AppShell.Main>

      <Drawer
        opened={ticketOpened}
        onClose={ticket.close}
        position="right"
        title="Place an order"
        size="md"
        overlayProps={{ backgroundOpacity: 0.55, blur: 3 }}
      >
        <OrderTicket defaultSymbol={activeSymbol} currency={data.activeMetadata?.currency} onComplete={completeOrder} />
      </Drawer>
      <DepositDialog
        opened={depositOpened}
        accounts={data.accounts}
        onClose={deposit.close}
      />
      <CreateAccountDialog opened={createAccountOpened} onClose={createAccount.close} />
    </AppShell>
  );
}
