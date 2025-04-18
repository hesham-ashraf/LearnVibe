# LearnVibe - Microservices-based Learning Platform

LearnVibe is an online learning platform built with a microservices architecture using Go.

## Project Overview

This project implements a learning platform with the following core services:

1. **User Management Service**: Handles user registration, authentication, and course enrollments.
2. **Content Management Service (CMS)**: Manages courses and course materials.
3. **API Gateway**: Routes requests to the appropriate microservices.

## Architecture

The platform follows a microservices architecture with:

- **API Gateway**: Routes requests to appropriate services
- **User Management Service**: Handles users and enrollments
- **CMS Service**: Manages courses and course materials
- **Shared Components**: Authentication, middleware, configuration utilities

## Features

### User Management
- User registration and login
- OAuth2 integration with Google
- JWT-based authentication
- Role-based access control (students, instructors, admins)
- Course enrollment management

### Content Management
- Course creation, editing, and deletion (instructors only)
- Course material management (PDF, video, links, text)
- Course listing and details
- Pagination support

### Security
- JWT token validation
- Role-based access control
- Password hashing

## Technologies

- **Go**: Core programming language
- **Gin**: Web framework
- **GORM**: ORM for database operations
- **PostgreSQL**: Primary database
- **Redis**: Caching
- **JWT**: Authentication tokens
- **OAuth2**: Google authentication



## License

[MIT License](LICENSE)
