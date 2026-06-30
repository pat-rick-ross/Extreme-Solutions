-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Customers table
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    address TEXT,
    package_id UUID,
    balance DECIMAL(10,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending',
    mikrotik_id VARCHAR(100),
    mikrotik_user VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    suspended_at TIMESTAMP,
    last_login_at TIMESTAMP
);

-- Packages table
CREATE TABLE packages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    speed VARCHAR(50) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    validity_days INT DEFAULT 30,
    is_active BOOLEAN DEFAULT true,
    bandwidth_up INT NOT NULL, -- Kbps
    bandwidth_down INT NOT NULL, -- Kbps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Invoices table
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    number VARCHAR(50) UNIQUE NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    tax DECIMAL(10,2) DEFAULT 0,
    total DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    description TEXT,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    due_date TIMESTAMP NOT NULL,
    paid_at TIMESTAMP,
    pdf_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Payments table
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id UUID REFERENCES invoices(id) ON DELETE SET NULL,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    method VARCHAR(20) NOT NULL,
    provider VARCHAR(20) NOT NULL,
    reference VARCHAR(100) UNIQUE,
    mpesa_receipt VARCHAR(50),
    mpesa_phone VARCHAR(20),
    description TEXT,
    metadata JSONB,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- MikroTik users table
CREATE TABLE mikrotik_users (
    id VARCHAR(100) PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    nas_id UUID,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    profile VARCHAR(100),
    uptime INT DEFAULT 0,
    bytes_in BIGINT DEFAULT 0,
    bytes_out BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- NAS devices table
CREATE TABLE nas_devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INT DEFAULT 8728,
    username VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    type VARCHAR(50) DEFAULT 'mikrotik',
    location VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    last_ping_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
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

-- Add foreign key for packages
ALTER TABLE customers ADD CONSTRAINT fk_customers_package FOREIGN KEY (package_id) REFERENCES packages(id) ON DELETE SET NULL;

-- Add foreign key for mikrotik_users
ALTER TABLE mikrotik_users ADD CONSTRAINT fk_mikrotik_users_nas FOREIGN KEY (nas_id) REFERENCES nas_devices(id) ON DELETE SET NULL;

-- Seed default packages
INSERT INTO packages (id, name, description, speed, price, validity_days, is_active, bandwidth_up, bandwidth_down) VALUES
    (uuid_generate_v4(), 'Starter', 'Basic internet package for light users', '10 Mbps', 1000, 30, true, 512, 1024),
    (uuid_generate_v4(), 'Standard', 'Medium package for home users', '20 Mbps', 2000, 30, true, 1024, 2048),
    (uuid_generate_v4(), 'Premium', 'High speed package for heavy users', '50 Mbps', 3500, 30, true, 2048, 5120),
    (uuid_generate_v4(), 'Ultimate', 'Ultimate speed for power users', '100 Mbps', 6000, 30, true, 5120, 10240);
