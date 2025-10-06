-- Create booking_services table
CREATE TABLE IF NOT EXISTS booking_services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    service_id UUID NOT NULL,
    stylist_id UUID NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Ensure end_time is after start_time
    CONSTRAINT chk_booking_services_time_order CHECK (end_time > start_time)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_booking_services_booking_id ON booking_services(booking_id);
CREATE INDEX IF NOT EXISTS idx_booking_services_service_id ON booking_services(service_id);
CREATE INDEX IF NOT EXISTS idx_booking_services_stylist_id ON booking_services(stylist_id);
CREATE INDEX IF NOT EXISTS idx_booking_services_start_time ON booking_services(start_time);
CREATE INDEX IF NOT EXISTS idx_booking_services_end_time ON booking_services(end_time);

-- Create composite index for stylist availability queries
CREATE INDEX IF NOT EXISTS idx_booking_services_stylist_time ON booking_services(stylist_id, start_time, end_time);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_booking_services_updated_at 
    BEFORE UPDATE ON booking_services 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
