# Booking Service

A comprehensive booking management microservice for the salon platform, handling the entire booking lifecycle from initiation to completion.

## Features

### Core Booking Lifecycle
- **Booking Initiation**: Create bookings with multiple services and stylists
- **Payment Integration**: Confirm bookings after successful payment (placeholder)
- **Booking Management**: View, reschedule, and cancel bookings
- **History Tracking**: Complete audit trail of all booking changes

### Availability Management
- **Real-time Slot Generation**: Dynamic availability based on stylist schedules
- **Buffer Time Management**: Configurable buffer between appointments
- **Conflict Detection**: Prevents double-booking and scheduling conflicts
- **Working Hours Integration**: Respects stylist schedules and breaks

### Pricing & Taxation
- **Dynamic Pricing**: Fetch service prices from salon-service
- **GST Calculation**: Configurable tax rates per branch
- **Booking Fees**: Branch-specific booking charges
- **Cost Breakdown**: Detailed pricing summary before payment

### Security & Authorization
- **JWT Authentication**: Customer token validation using salon-shared middleware
- **User-Scoped Access**: Users can only access their own bookings
- **Audit Logging**: Comprehensive logging of all booking operations
- **Rate Limiting**: Protection against abuse

## Architecture

```
booking-service/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── api/
│   │   └── handlers.go         # HTTP request handlers
│   ├── config/
│   │   └── config.go          # Configuration management
│   ├── db/
│   │   ├── connection.go      # Database connection
│   │   └── migrations.go      # Migration runner
│   ├── model/
│   │   ├── booking.go         # Core booking models
│   │   └── enums.go          # Status enums
│   ├── repository/
│   │   └── booking_repository.go # Data access layer
│   └── service/
│       ├── booking_service.go     # Business logic
│       └── external_services.go  # External API integration
├── migrations/
│   ├── 001_create_bookings_table.sql
│   ├── 002_create_booking_services_table.sql
│   ├── 003_create_booking_history_table.sql
│   └── 004_create_branch_configurations_table.sql
└── configs/
    └── config.yaml           # Default configuration
```

## Database Schema

### Core Tables

#### `bookings`
- Primary booking entity with user, salon, branch references
- Status tracking (initiated, confirmed, rescheduled, canceled, completed)
- Payment information and totals
- Audit timestamps

#### `booking_services`
- Individual services within a booking
- Stylist assignments and time slots
- Service pricing and duration

#### `booking_history`
- Complete audit trail of booking changes
- JSON storage of old/new values
- User attribution and reasoning

#### `branch_configurations`
- Branch-specific booking settings
- Buffer times, cancellation policies
- Pricing configuration (GST, booking fees)

## API Endpoints

### Booking Management
```http
POST   /api/v1/bookings/initiate           # Create new booking
POST   /api/v1/bookings/confirm            # Confirm after payment
GET    /api/v1/bookings/{id}               # Get booking details
PATCH  /api/v1/bookings/{id}/cancel        # Cancel booking
PATCH  /api/v1/bookings/{id}/reschedule    # Reschedule booking
```

### User Bookings
```http
GET    /api/v1/bookings/user/{userId}      # Get user's bookings
```

### Availability & Pricing
```http
GET    /api/v1/stylists/{id}/availability  # Get available slots
POST   /api/v1/bookings/summary            # Calculate pricing
```

### Configuration
```http
GET    /api/v1/branches/{id}/config        # Get branch settings
```

### Health & Monitoring
```http
GET    /health                             # Health check
GET    /ready                              # Readiness check
```

## Request/Response Examples

### Initiate Booking
```json
POST /api/v1/bookings/initiate
{
  "salon_id": "123e4567-e89b-12d3-a456-426614174000",
  "branch_id": "123e4567-e89b-12d3-a456-426614174001",
  "services": [
    {
      "service_id": "123e4567-e89b-12d3-a456-426614174002",
      "stylist_id": "123e4567-e89b-12d3-a456-426614174003",
      "start_time": "2024-01-15T10:00:00Z"
    }
  ],
  "notes": "First time customer"
}
```

### Booking Response
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174004",
  "user_id": "123e4567-e89b-12d3-a456-426614174005",
  "salon_id": "123e4567-e89b-12d3-a456-426614174000",
  "branch_id": "123e4567-e89b-12d3-a456-426614174001",
  "status": "initiated",
  "total_amount": 1180.00,
  "gst": 180.00,
  "booking_fee": 50.00,
  "payment_status": "pending",
  "services": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174006",
      "service_id": "123e4567-e89b-12d3-a456-426614174002",
      "stylist_id": "123e4567-e89b-12d3-a456-426614174003",
      "start_time": "2024-01-15T10:00:00Z",
      "end_time": "2024-01-15T11:00:00Z",
      "price": 1000.00
    }
  ],
  "created_at": "2024-01-10T09:00:00Z",
  "updated_at": "2024-01-10T09:00:00Z"
}
```

### Availability Response
```json
GET /api/v1/stylists/{id}/availability?date=2024-01-15
{
  "stylist_id": "123e4567-e89b-12d3-a456-426614174003",
  "date": "2024-01-15",
  "slots": [
    {
      "start_time": "2024-01-15T09:00:00Z",
      "end_time": "2024-01-15T09:30:00Z",
      "available": true
    },
    {
      "start_time": "2024-01-15T09:30:00Z",
      "end_time": "2024-01-15T10:00:00Z",
      "available": true
    }
  ]
}
```

## Configuration

### Environment Variables
```bash
# Service Configuration
BOOKING_SERVICE_ENV=development
BOOKING_SERVICE_PORT=8082
BOOKING_SERVICE_DB_URL=postgres://user:pass@localhost:5432/salon

# JWT Configuration
BOOKING_SERVICE_JWT_ACCESSSECRET=your-access-secret
BOOKING_SERVICE_JWT_REFRESHSECRET=your-refresh-secret

# External Services
USER_SERVICE_URL=http://localhost:8080
SALON_SERVICE_URL=http://localhost:8081

# Default Booking Settings
DEFAULT_BUFFER_TIME_MINUTES=15
DEFAULT_CANCELLATION_CUTOFF_HOURS=2
DEFAULT_RESCHEDULE_WINDOW_HOURS=4
DEFAULT_MAX_ADVANCE_BOOKING_DAYS=30
DEFAULT_BOOKING_FEE_AMOUNT=50.0
DEFAULT_GST_PERCENTAGE=18.0
```

### Branch Configuration
Each branch can have custom settings that override defaults:
- **Buffer Time**: Minutes between appointments
- **Cancellation Policy**: Hours before appointment
- **Reschedule Window**: Hours before appointment
- **Advance Booking**: Maximum days in advance
- **Pricing**: Booking fees and GST rates

## External Service Integration

### User Service
- **User Validation**: Verify customer exists and is active
- **Profile Information**: Get user details for notifications

### Salon Service
- **Salon/Branch Data**: Validate salon and branch information
- **Service Information**: Get service details, pricing, duration
- **Stylist Management**: Validate stylists and get schedules
- **Working Hours**: Fetch stylist availability and breaks

### Payment Service (Placeholder)
- **Payment Processing**: Handle payment transactions
- **Refund Management**: Process cancellation refunds

### Notification Service (Placeholder)
- **Booking Confirmations**: Send confirmation messages
- **Reminders**: Appointment reminder notifications
- **Status Updates**: Reschedule and cancellation notices

## Business Rules

### Booking Validation
- Users can only book for themselves
- Services must be available at selected branch
- Stylists must be qualified for selected services
- Appointment times must be in the future
- Buffer time must be respected between appointments

### Cancellation Policy
- Bookings can be canceled up to configured cutoff time
- Refunds processed automatically for valid cancellations
- History maintained for all cancellation reasons

### Rescheduling Rules
- Must be within reschedule window
- New time slots must be available
- Same validation rules as new bookings
- Original booking marked as rescheduled

### Pricing Calculation
```
Subtotal = Sum of all service prices
GST = Subtotal × (GST Percentage / 100)
Total = Subtotal + Booking Fee + GST
```

## Development

### Prerequisites
- Go 1.23+
- PostgreSQL 13+
- Access to salon-shared module

### Local Setup
```bash
# Clone repository
git clone <repository-url>
cd salon/booking-service

# Install dependencies
go mod tidy

# Set up database
createdb salon
export BOOKING_SERVICE_DB_URL="postgres://user:pass@localhost:5432/salon?sslmode=disable"

# Run migrations
go run cmd/main.go

# Start service
go run cmd/main.go
```

### Docker Setup
```bash
# Build image
docker build -f ../Dockerfile.booking-service -t booking-service .

# Run container
docker run -p 8082:8082 \
  -e BOOKING_SERVICE_DB_URL="postgres://user:pass@host:5432/salon" \
  -e BOOKING_SERVICE_JWT_ACCESSSECRET="your-secret" \
  booking-service
```

## Testing

### API Testing
```bash
# Health check
curl http://localhost:8082/health

# Initiate booking (requires JWT token)
curl -X POST http://localhost:8082/api/v1/bookings/initiate \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{"salon_id":"...","branch_id":"...","services":[...]}'
```

### Load Testing
- Use tools like Apache Bench or k6 for performance testing
- Focus on availability calculation endpoints
- Test concurrent booking scenarios

## Monitoring & Observability

### Health Checks
- `/health`: Basic service health
- `/ready`: Database connectivity and readiness

### Logging
- Structured JSON logging with zerolog
- Request/response logging via middleware
- Business event logging (bookings, cancellations)

### Metrics (Future)
- Booking success/failure rates
- Average booking value
- Popular time slots and services
- Cancellation rates by reason

## Security Considerations

### Authentication
- JWT tokens validated using salon-shared middleware
- Customer-type tokens required for all endpoints
- User-scoped access control

### Data Protection
- Sensitive data sanitized in logs
- Audit trail for all booking changes
- Rate limiting on booking operations

### Input Validation
- Comprehensive request validation
- SQL injection prevention via parameterized queries
- XSS protection through proper encoding

## Future Enhancements

### Advanced Features
- **Recurring Bookings**: Weekly/monthly appointment scheduling
- **Group Bookings**: Multiple customers in single booking
- **Waitlist Management**: Automatic rebooking when slots open
- **Dynamic Pricing**: Peak hour and demand-based pricing

### Integration Improvements
- **Real-time Notifications**: WebSocket-based updates
- **Calendar Integration**: Sync with external calendars
- **SMS/Email Templates**: Rich notification formatting
- **Payment Gateway**: Multiple payment method support

### Analytics & Reporting
- **Booking Analytics**: Trends and patterns analysis
- **Revenue Reporting**: Financial insights and forecasting
- **Customer Insights**: Booking behavior analysis
- **Operational Metrics**: Efficiency and utilization reports

## Support

For issues and questions:
1. Check the logs for error details
2. Verify external service connectivity
3. Validate database schema and migrations
4. Review configuration settings

Common troubleshooting:
- **Database Connection**: Check connection string and credentials
- **JWT Errors**: Verify token format and shared secrets
- **External Service**: Confirm user-service and salon-service availability
- **Migration Issues**: Ensure database permissions and schema access
