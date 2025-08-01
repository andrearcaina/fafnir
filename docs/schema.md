# Initial Schema Design

```sql
-- stores user information
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    hashed_password VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- stores role information
CREATE TABLE roles (
    name VARCHAR(50) PRIMARY KEY,
    description TEXT
);

-- stores permission information
CREATE TABLE permissions (
    name VARCHAR(50) PRIMARY KEY,
    description TEXT
);

-- links roles to permissions
CREATE TABLE roles_permissions (
    role_name VARCHAR(50) REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

-- links users to roles
CREATE TABLE users_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_name VARCHAR(50) REFERENCES roles(name) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_name)
);

-- stores account information for users (like their balance)
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    balance NUMERIC NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- stores stock information
CREATE TABLE stock_metadata (
    symbol VARCHAR(10) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    type VARCHAR(50) NOT NULL,
    sector VARCHAR(100) NOT NULL
);

-- This table stores the latest stock quote information (data gets overwritten if updated_at > 60 seconds, acts like caching)
CREATE TABLE stock_quote (
    symbol VARCHAR(10) REFERENCES stock_metadata(symbol) ON DELETE CASCADE,
    last_price NUMERIC NOT NULL,
    price_change NUMERIC NOT NULL,
    price_change_pct NUMERIC NOT NULL,
    volume BIGINT NOT NULL,
    market_cap BIGINT NOT NULL,
    pe_ratio NUMERIC NOT NULL,
    dividend_yield NUMERIC NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- stores user portfolio information (a list of stocks a user holds)
CREATE TABLE portfolio (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID REFERENCES accounts(id) ON DELETE CASCADE,
    symbol VARCHAR(10) REFERENCES stock_metadata(symbol) ON DELETE CASCADE,
    quantity NUMERIC(20,6) NOT NULL CHECK (quantity >= 0),
    avg_buy_price NUMERIC(20,6) NOT NULL CHECK (avg_buy_price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- stores user watchlist information (a list of stocks a user is interested in, but not necessarily holding)
CREATE TABLE watchlist (
   user_id UUID REFERENCES users(id) ON DELETE CASCADE,
   symbol TEXT REFERENCES stock_metadata(symbol),
   added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   PRIMARY KEY (user_id, symbol)
);
```