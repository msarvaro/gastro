-- Create businesses table
CREATE TABLE IF NOT EXISTS businesses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    address VARCHAR(255),
    phone VARCHAR(50),
    email VARCHAR(255),
    website VARCHAR(255),
    logo VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Add business_id to users table
ALTER TABLE users
ADD COLUMN business_id INTEGER REFERENCES businesses(id) ON DELETE CASCADE;

-- Add business_id to tables table
ALTER TABLE tables
ADD COLUMN business_id INTEGER REFERENCES businesses(id) ON DELETE CASCADE;

-- Add business_id to categories table
ALTER TABLE categories
ADD COLUMN business_id INTEGER REFERENCES businesses(id) ON DELETE CASCADE;

-- Add business_id to dishes table
ALTER TABLE dishes
ADD COLUMN business_id INTEGER REFERENCES businesses(id) ON DELETE CASCADE;

-- Add business_id to orders table
ALTER TABLE orders
ADD COLUMN business_id INTEGER REFERENCES businesses(id) ON DELETE CASCADE;

-- Add business_id to order_items table
ALTER TABLE order_items
ADD COLUMN business_id INTEGER REFERENCES businesses(id) ON DELETE CASCADE;

-- Add business_id to inventory table
ALTER TABLE inventory
ADD COLUMN business_id INTEGER REFERENCES businesses(id) ON DELETE CASCADE;

-- Add business_id to suppliers table
ALTER TABLE suppliers
ADD COLUMN business_id INTEGER REFERENCES businesses(id) ON DELETE CASCADE;

-- Add business_id to requests table
ALTER TABLE requests
ADD COLUMN business_id INTEGER REFERENCES businesses(id) ON DELETE CASCADE;

-- Add business_id to shifts table
ALTER TABLE shifts
ADD COLUMN business_id INTEGER REFERENCES businesses(id) ON DELETE CASCADE;

-- Add business_id to shift_employees table
ALTER TABLE shift_employees
ADD COLUMN business_id INTEGER REFERENCES businesses(id) ON DELETE CASCADE; 