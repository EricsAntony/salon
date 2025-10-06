-- Create bookings table
CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    salon_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('initiated', 'confirmed', 'rescheduled', 'canceled', 'completed')),
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    gst DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    booking_fee DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded')),
    payment_id VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_salon_id ON bookings(salon_id);
CREATE INDEX IF NOT EXISTS idx_bookings_branch_id ON bookings(branch_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
CREATE INDEX IF NOT EXISTS idx_bookings_payment_status ON bookings(payment_status);
CREATE INDEX IF NOT EXISTS idx_bookings_created_at ON bookings(created_at);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_bookings_updated_at 
    BEFORE UPDATE ON bookings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
