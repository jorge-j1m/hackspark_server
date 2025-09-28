# HackSpark Server

The backend API server for HackSpark, a platform designed to inspire developers to start new projects by providing personalized project suggestions, community engagement features, and gamified progress tracking.

## What it does

HackSpark Server is a RESTful API built with Go that powers the HackSpark platform. It provides the core functionality needed to support a community of developers working on personal projects, including:

**Core Features:**
- **User Authentication & Management** - Secure user registration, login, and profile management with bcrypt password hashing
- **Project Management** - Full CRUD operations for projects with like/unlike functionality and engagement tracking
- **Technology Tags** - Categorization system for projects and user skills using technology tags
- **Community Features** - User profiles, project discovery, and social interaction through likes and follows
- **Trending Analytics** - Leaderboards and trending insights based on project engagement

When users interact with the platform, they can create detailed profiles specifying their technical skills, browse and discover projects by technology stack, engage with the community through likes and follows, and track their progress through the gamified point system.

## How we built it

HackSpark Server is architected as a modern, production-ready Go application following clean architecture principles and industry best practices:

**Technology Stack:**
- **Go 1.25.1** - Modern, performant backend language
- **Ent Framework** - Type-safe ORM with automatic code generation for database operations
- **Chi Router** - Lightweight, fast HTTP router with middleware support
- **PostgreSQL 17** - Robust relational database for data persistence
- **Docker & Docker Compose** - Containerized deployment with orchestration
- **Zerolog** - Structured logging for observability
- **TypeID** - Type-safe, globally unique identifiers

**Architecture Highlights:**
- **Clean Architecture** - Separation of concerns with distinct layers for infrastructure, interfaces, and business logic
- **Database Schema Design** - Well-normalized schema with proper indexing and relationships using Ent's type-safe approach
- **Middleware Pipeline** - Comprehensive middleware for authentication, CORS, logging, security headers, and request tracing
- **Production Deployment** - Multi-stage Docker builds using distroless images for minimal attack surface

The codebase structure follows Go conventions with clear separation between:
```
cmd/              # Application entry points
internal/         # Private application code
├── infrastructure/  # Config, logging, server setup
├── interfaces/     # HTTP handlers, middleware, routing
└── pkg/           # Shared utilities and common errors
ent/              # Generated database code and schemas
```

## Technical Implementation Details

**Database Design:**
The system uses a carefully designed relational schema with seven core entities:
- **Users** - Complete user profiles with authentication, verification, and account management
- **Projects** - Project data with ownership, engagement metrics, and tagging
- **Sessions** - Secure session management for authentication
- **Tags** - Technology categorization system
- **Likes** - User-project engagement tracking
- **UserTechnology** - Many-to-many relationship for user skills
- **ProjectTag** - Many-to-many relationship for project categorization

**API Design:**
RESTful API following OpenAPI conventions with versioned endpoints (`/api/v1/`):
- `/auth/*` - Authentication endpoints (signup, login, logout)
- `/users/*` - User management and profile operations
- `/projects/*` - Project CRUD and engagement features
- `/tags/*` - Technology tag management and trending data

**Security & Production Readiness:**
- Password hashing with bcrypt and automatic hooks
- CORS configuration for cross-origin requests
- Security headers middleware for protection against common attacks
- Request timeout and rate limiting
- Structured logging with request tracing
- Health check endpoints for monitoring
- Distroless Docker images for minimal attack surface

## Challenges we ran into

Building a production-ready backend system presented several technical challenges:

**Database Complexity:** Designing a flexible schema that could handle the many-to-many relationships between users, projects, and tags while maintaining query performance required careful consideration of indexing strategies and join optimizations.

**Type Safety at Scale:** Implementing the Ent framework for type-safe database operations involved learning its code generation patterns and hook system, particularly for complex operations like automatic password hashing and relationship management.

**Authentication Architecture:** Building a secure, stateless authentication system that could scale across multiple services while maintaining session management and user verification workflows.

**Production Deployment:** Configuring a robust containerized deployment with proper health checks, logging, and monitoring while maintaining security best practices through distroless images and non-root users.

## Accomplishments that we're proud of

**Code Quality and Maintainability:** The codebase demonstrates enterprise-level patterns with clean architecture, comprehensive error handling, and extensive use of Go's type system for compile-time safety.

**Database Schema Excellence:** The Ent-powered schema provides full type safety from database to API, with automatic code generation, migration management, and relationship handling that scales efficiently.

**Production-Ready Infrastructure:** The deployment setup includes multi-stage Docker builds, health monitoring, structured logging, and security hardening that would be suitable for production environments.

**Performance Optimization:** Strategic use of database indexing, efficient query patterns, and proper middleware ordering ensure the API can handle significant load while maintaining responsiveness.

## What we learned

**Advanced Go Patterns:** Deep dive into Go's interface system, context propagation, and error handling patterns that enable building robust, concurrent systems.

**ORM and Database Design:** Hands-on experience with modern ORM frameworks, understanding the tradeoffs between type safety and flexibility, and designing schemas that perform well at scale.

**Production System Architecture:** Real-world experience with containerization, security hardening, monitoring, and deployment patterns that bridge the gap between development and production.

**API Design Philosophy:** Understanding REST principles, versioning strategies, and how to design APIs that are both developer-friendly and performance-optimized.

## What's next for HackSpark Server

**Scalability Enhancements:** Implementation of caching layers (Redis), database read replicas, and connection pooling to handle increased user load and data volume.

**Advanced Features:** Integration with external APIs for project suggestion algorithms, webhook systems for real-time notifications, and analytics pipeline for usage insights.

**Monitoring and Observability:** Addition of metrics collection (Prometheus), distributed tracing (Jaeger), and comprehensive alerting to ensure system reliability at scale.

**Security Hardening:** Implementation of rate limiting, API key management, audit logging, and enhanced authentication features like multi-factor authentication and OAuth integration.

The server architecture provides a solid foundation that can evolve from supporting a small community to handling enterprise-scale deployments while maintaining code quality and developer productivity.

## Getting Started

### Prerequisites
- Go 1.25.1+
- Docker and Docker Compose
- PostgreSQL 17 (or use the provided Docker setup)

### Quick Start
```bash
# Clone the repository
git clone https://github.com/jorge-j1m/hackspark_server
cd hackspark_server

# Start the services using Docker Compose
docker-compose up -d

# The API will be available at http://localhost:8080
# Health check: http://localhost:8080/health
```

### Development Setup
```bash
# Install dependencies
go mod download

# Set up environment variables
cp .env.example .env

# Run database migrations (with Ent)
go generate ./...

# Start the development server
go run cmd/api/main.go
```

### API Documentation
API endpoints are available at `/api/v1/` with comprehensive error handling and validation. See the Insomnia collection (`Insomnia_2025-09-27.yaml`) for complete API documentation and examples.