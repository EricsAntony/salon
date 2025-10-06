-- Create booking_history table
CREATE TABLE IF NOT EXISTS booking_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL CHECK (action IN ('created', 'confirmed', 'rescheduled', 'canceled', 'completed')),
    old_values JSONB,
    new_values JSONB,
    user_id UUID,
    reason TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_booking_history_booking_id ON booking_history(booking_id);
CREATE INDEX IF NOT EXISTS idx_booking_history_action ON booking_history(action);
CREATE INDEX IF NOT EXISTS idx_booking_history_timestamp ON booking_history(timestamp);
CREATE INDEX IF NOT EXISTS idx_booking_history_user_id ON booking_history(user_id);
