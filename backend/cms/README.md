# Course Management Service (CMS)

This microservice is part of the LearnVibe learning platform, focused on course management and user enrollment. 

## Features

- **Course Management**: Create, read, update, and delete courses
- **Authentication**: OAuth2 with Google and JWT-based authentication
- **Authorization**: Role-based access control ensuring only instructors and admins can manage courses
- **Content Management**: Add various types of content to courses (PDF, videos, links, text)
- **User Enrollment**: Allow students to enroll in courses and track their progress

## API Endpoints

### Authentication

- `GET /auth/google`: Initiates Google OAuth2 login
- `GET /auth/google/callback`: Callback URL for Google OAuth2

### Courses

- `GET /api/courses`: List all courses with pagination
- `GET /api/courses/:id`: Get a specific course by ID
- `POST /api/courses`: Create a new course (instructors/admins only)
- `PUT /api/courses/:id`: Update a course (instructors/admins only)
- `DELETE /api/courses/:id`: Delete a course (instructors/admins only)
- `POST /api/courses/:id/contents`: Add content to a course (instructors/admins only)
- `DELETE /api/courses/:id/contents/:contentId`: Delete content from a course (instructors/admins only)
- `GET /api/courses/:id/enrollments`: List all enrollments for a course (instructors/admins only)

### Enrollments

- `POST /api/courses/:id/enroll`: Enroll in a course
- `GET /api/enrollments`: List all courses a user is enrolled in
- `GET /api/enrollments/:id`: Get details of a specific enrollment
- `PUT /api/enrollments/:id/progress`: Update progress in a course
- `PUT /api/enrollments/:id/drop`: Drop a course

## Getting Started

### Prerequisites

- Go 1.21 or later
- PostgreSQL database

### Database Setup

1. Create a PostgreSQL database named `learnvibe`
2. Use the following connection details:
   - Username: postgres
   - Password: vampire8122003 (change this in production)
   - Host: localhost
   - Port: 5432
   - Database: learnvibe

3. Run the application to automatically migrate database tables:
   ```bash
   go run main.go
   ```

4. You can verify the database tables by running:
   ```bash
   go run scripts/check_db.go
   ```

### Environment Variables

Set the following environment variables:

- `PORT`: Server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for JWT signing
- `GOOGLE_CLIENT_ID`: Google OAuth2 client ID
- `GOOGLE_CLIENT_SECRET`: Google OAuth2 client secret
- `GOOGLE_REDIRECT_URL`: OAuth2 callback URL (default: http://localhost:8080/auth/google/callback)

### Running locally

```bash
# Install dependencies
go mod download

# Run the application
go run main.go
```

## Development

### Testing

```bash
go test ./... -v
```

### Creating an admin user

By default, all users are created with the "student" role. To create an admin:

1. Connect to the database
2. Update a user's role:
   ```sql
   UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';
   ``` 