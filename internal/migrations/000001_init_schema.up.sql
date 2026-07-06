-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. PACKAGES TABLE
CREATE TABLE packages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    speed VARCHAR(50) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    validity_days INT DEFAULT 30,
    is_active BOOLEAN DEFAULT true,
    bandwidth_up INT NOT NULL,   -- Explicitly tracked in Kbps
    bandwidth_down INT NOT NULL, -- Explicitly tracked in Kbps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. CUSTOMERS TABLE (Cleaned of redundant MikroTik data)
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL, -- Keep in E.164 format (+254...)
    address TEXT,
    package_id UUID REFERENCES packages(id) ON DELETE SET NULL,
    balance DECIMAL(10,2) DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'pending', -- pending, active, suspended, terminated
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    suspended_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- 3. INVOICES TABLE
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    number VARCHAR(50) UNIQUE NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    tax DECIMAL(10,2) DEFAULT 0.00,
    total DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- unpaid, paid, partially_paid, cancelled
    description TEXT,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    due_date TIMESTAMP WITH TIME ZONE NOT NULL,
    paid_at TIMESTAMP WITH TIME ZONE,
    pdf_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. PAYMENTS TABLE
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id UUID REFERENCES invoices(id) ON DELETE SET NULL,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, completed, failed
    method VARCHAR(20) NOT NULL,          -- mpesa, bank, cash
    provider VARCHAR(20) NOT NULL,        -- safaricom, equity
    reference VARCHAR(100) UNIQUE,        -- External Trans ID (e.g., SBR123XYZ)
    mpesa_receipt VARCHAR(50),
    mpesa_phone VARCHAR(20),
    description TEXT,
    metadata JSONB,                       -- Clean storage for raw callback JSON hooks
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. NAS DEVICES TABLE
CREATE TABLE nas_devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    host INET NOT NULL,                   -- Swapped string for native INET structure
    port INT DEFAULT 8728,
    username VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    type VARCHAR(50) DEFAULT 'mikrotik',
    location VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    last_ping_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 6. MIKROTIK USERS TABLE
CREATE TABLE mikrotik_users (
    id VARCHAR(100) PRIMARY KEY,          -- Stays string if holding RouterOS internal IDs (*1, *2)
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    nas_id UUID REFERENCES nas_devices(id) ON DELETE SET NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    profile VARCHAR(100),
    ip_address INET UNIQUE,               -- Native tracking for static subnetting/pools
    uptime INT DEFAULT 0,
    bytes_in BIGINT DEFAULT 0,
    bytes_out BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- --- ENHANCED INDEXES FOR AUTOMATION LOOKUPS ---
CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_phone ON customers(phone);
CREATE INDEX idx_customers_status ON customers(status);
CREATE INDEX idx_invoices_customer_id ON invoices(customer_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);
CREATE INDEX idx_payments_customer_id ON payments(customer_id);
CREATE INDEX idx_payments_invoice_id ON payments(invoice_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_mikrotik_users_customer_id ON mikrotik_users(customer_id);
CREATE INDEX idx_mikrotik_users_username ON mikrotik_users(username);
CREATE INDEX idx_mikrotik_users_ip ON mikrotik_users(ip_address);

-- --- SEED CORRECTED PACKAGE BANDWIDTHS (10 Mbps = 10240 Kbps) ---
INSERT INTO packages (id, name, description, speed, price, validity_days, is_active, bandwidth_up, bandwidth_down) VALUES
    (uuid_generate_v4(), 'Starter', 'Basic internet package for light users', '10 Mbps', 1000.00, 30, true, 10240, 10240),
    (uuid_generate_v4(), 'Standard', 'Medium package for home users', '20 Mbps', 2000.00, 30, true, 20480, 20480),
    (uuid_generate_v4(), 'Premium', 'High speed package for heavy users', '50 Mbps', 3500.00, 30, true, 51200, 51200),
    (uuid_generate_v4(), 'Ultimate', 'Ultimate speed for power users', '100 Mbps', 6000.00, 30, true, 102400, 102400);