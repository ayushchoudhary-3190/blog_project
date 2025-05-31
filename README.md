# README.md

# Blog Application

This project is a blog application built with Go. It includes features such as user authentication, session management, and the ability to create and manage blog posts.

## Features

- User authentication (sign up and sign in)
- Session management
- Create, read, and manage blog posts

## Setup Instructions

1. **Clone the repository:**
   ```
   git clone <repository-url>
   cd blog_app
   ```

2. **Install dependencies:**
   Make sure you have Go installed on your machine. Then, run:
   ```
   go mod tidy
   ```

3. **Set up the database:**
   Ensure you have MySQL running and create two databases: `users` and `blogs`. Update the database connection strings in `cmd/main.go` if necessary.

4. **Run the application:**
   You can run the application using:
   ```
   go run cmd/main.go
   ```

5. **Development with Air:**
   For live reloading during development, you can use the Air tool. Make sure to have it installed, then run:
   ```
   air
   ```

## Usage

- Navigate to `http://localhost:8080` to access the application.
- You can sign up for a new account or sign in if you already have one.
- Once logged in, you can create new blog posts.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
