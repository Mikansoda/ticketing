# About the project: A backend system to manage inventory management
This project is a RESTful API built with Golang (Gin Framework) that manages inventory and sales transactions of an e-commerce.
It includes features such as CRUD operations, authentication with JWT, role-based access control, middleware integration, and deployment to VPS.

# Tech stack
1. Language: Golang 1.24.2
2. Framework: Gin Web Framework
3. ORM: GORM
4. Database: MySQL
5. Authentication: JWT

# Database Design
The database is designed with multiple entities:
1.  Users
    Stores user information including username, email, password hash,  role (admin or user), and account status. Users can have multiple addresses.
2.  Visitors
    This represents personal details of a user’s visitor, including a unique national ID, name, contact number, and nationality. Each Visitor is tied to a User via `BuyerID`.
3.  Events
    Contains details about the events, including the name, category, location, and date. It also has relations to Event Categories and Ticket Types.
4.  Event Categories
    Categorizes events into different groups (e.g., concert, sports), which can aid in filtering (used in GET all event query params).
5.  Event Images
    Stores image urls of events, including whether the image is marked as primary for event promotions.
6.  Ticket Types
    Represents different types of tickets for events (e.g., VIP, Regular), including price, quota, and status. It has a relation to the Events table, showing which event the ticket belongs to.
7.  Bookings
    Represents the booking transaction. It connects a User with Tickets, showing the quantity, status, and total cost of tickets purchased.
8.  Tickets
    Actual ticket data, used for check-ins and is linked to a booking, the type of ticket, and the visitor attending. Includes status (pending, used, etc.) and seat info.
9.  Payments
    Stores payment information for orders, including invoice ID, payment type, and status. Each payment belongs to one order.

# Installation and Setup

1. Clone the repository
2. Create a database, sql script provided in this repository
3. Create an `.env` (example included in this repository) file which includes:
```
APP_PORT=8080

DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_db_password
DB_NAME=ticketing

JWT_ACCESS_SECRET=your_jwt_access_secret
JWT_REFRESH_SECRET=your_jwt_refresh_secret
ACCESS_TTL_MIN=15
REFRESH_TTL_DAYS=7

SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=your_email@example.com
SMTP_PASS=your_email_password
FROM_EMAIL=your_email@example.com
APP_ENV=dev

CLOUDINARY_URL=your_cloudinary_url
XENDIT_API_KEY=your_xendit_api_key
```

4. Run the server using `go run main.go` on the terminal

# Documentation
## Endpoints
### 1. Authentication
- `POST /auth/register` – Register a new user
- `POST /auth/login` – Login user and receive JWT tokens
- `POST /auth/verify-otp` – Verify OTP for account activation
- `POST /auth/refresh` – Refresh JWT tokens
- `POST /auth/logout` – Logout user
- `GET /auth/profile` – Get current user profile (user or admin)
- `GET /auth/admin/dashboard` – Admin-only dashboard

### 2. Events
- `GET /events` – Get list of events (supports search, category filter, pagination)
- `GET /events/:eventId` – Get event by ID
- `POST /events` – Create an event (admin only)
- `PATCH /admin/events/:eventId` – Update an event (admin only)
- `DELETE /admin/events/:eventId` – Delete an event (admin only)
- `POST /admin/events/:eventId/recover` – Recover a deleted event (admin only)
- `GET /categories` – Get list of category (supports pagination)
- `POST /admin/categories` – Create a category (admin only)
- `PATCH /admin/categories/:id` – Update a category (admin only)
- `DELETE /admin/categories/:id` – Delete a category (admin only)
- `POST /admin/categories/:id/recover` – Recover a deleted category (admin only)
- `POST /admin/events/:eventId/images` – Upload event image (admin only)
- `DELETE /admin/images/:imageId` – Delete event image (admin only)
- `POST /admin/images/:imageId/recover` – Recover deleted image (admin only)

### 3. Ticket Types
- `GET /ticket-types` – Get list of ticket types (supports search, pagination)
- `GET /ticket-types/:id` – Get ticket types by ID
- `POST /admin/ticket-types` – Create a ticket types (admin only)
- `PATCH /admin/ticket-types/:id` – Update a ticket types (admin only)
- `DELETE /admin/ticket-types/:id` – Delete a ticket types (admin only)
- `POST /admin/ticket-types/:idrecover` – Recover a deleted ticket types (admin only)

### 4. Bookings & Tickets
- `GET /user/bookings` – Get all user’s own bookings without ticket details, excluding visitor data (user only)
- `GET /bookings/:id` – Get a user’s bookings with ticket details, including visitor data (user/admin only)
- `GET /user/tickets/:ticketId` – Get ticket by ID (user only)
- `POST /user/bookings` – Create new booking (user only, rate limit: 10/min)
- `GET /admin/bookings` – Get all user’s own bookings without ticket details, excluding visitor data (admin only)
- `PATCH /admin/bookings/:id/status` – Update booking status (manually mainly for refund reasons) (admin only)
- `PATCH /admin/tickets/:ticketId/status` – Update ticket status (manually mainly for check-in reasons) (admin only)

### 6. Payments
- `POST /user/payments/xendit` – Create payment via Xendit (user only, rate-limited)
- `GET /user/payments` – Get user's payments (user only, rate-limited)
- `GET /admin/payments` – Get all payments (admin only)
- `POST /admin/payments/webhook/xendit` – Xendit payment webhook (admin only)

### 7. Reports
- `GET /admin/reports/summary` – View revenue and sold ticket count in a certain period
- `GET /admin/event/:id` – View revenue and sold ticket count in a certain event

**Note:**  
Some routes are protected by JWT authentication (user or admin), some has rate limits to prevent abuse (e.g., booking actions, image uploads, payments), or both.

**Author:**<br>
Developed by Zahra<br>
Final Project for Dibimbing Bootcamp - Golang Back-End development