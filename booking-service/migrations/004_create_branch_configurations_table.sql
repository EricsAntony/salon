-- Create branch_configurations table
CREATE TABLE IF NOT EXISTS branch_configurations (
    branch_id UUID PRIMARY KEY,
    buffer_time_minutes INTEGER NOT NULL DEFAULT 15 CHECK (buffer_time_minutes >= 0),
    cancellation_cutoff_hours INTEGER NOT NULL DEFAULT 2 CHECK (cancellation_cutoff_hours >= 0),
    reschedule_window_hours INTEGER NOT NULL DEFAULT 4 CHECK (reschedule_window_hours >= 0),
    max_advance_booking_days INTEGER NOT NULL DEFAULT 30 CHECK (max_advance_booking_days > 0),
    booking_fee_amount DECIMAL(10,2) NOT NULL DEFAULT 50.00 CHECK (booking_fee_amount >= 0),
    gst_percentage DECIMAL(5,2) NOT NULL DEFAULT 18.00 CHECK (gst_percentage >= 0 AND gst_percentage <= 100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_branch_configurations_updated_at 
    BEFORE UPDATE ON branch_configurations 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default configuration for existing branches (if any)
-- This will be handled by the application when branches are created
