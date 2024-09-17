# Memorization Tracker API

This is a RESTful API for tracking Quran memorization progress, implemented using the GIN framework and GORM for database handling.

## Getting Started

### Prerequisites

- Go 1.19 or later
- PostgreSQL database
- GIN framework
- GORM library

### Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/radityafijarp/be-quran-app-final-project.git
    cd your-repository
    ```

2. Install the required Go modules:
    ```bash
    go mod tidy
    ```

3. Set up your PostgreSQL database and modify the `dbCredential` in `main.go` to match your database credentials.

### Running the Server

To run the server, use the following command:

```bash
go run main.go
```
This will start the API server at http://localhost:8080.

### Running Test Code
To run the tests for the project, execute the following command:
```bash
go test 
```
This will run all the test files located in your project.

### Endpoints and Usage
#### Authentication Endpoints
1. Register a User
    - Endpoint: POST /users
    - Request Body:

      ```bash
      {
        "username": "john_doe",
        "password": "password123"
      }
      ```
    - Response: Status 201 Created or 409 Conflict if the username is already registered.
2. Login User
    - Endpoint: POST /users
    - Request Body:

      ```bash
      {
        "username": "john_doe",
        "password": "password123"
      }
      ```
    - Response: Status 200 OK with JWT token.

#### Memorization Endpoints
Authenticated requests to the following endpoints must include a JWT token in the Authorization header like so:

  ```bash
  Authorization: Bearer <your-token>
  ```
1. Get All Memorizes
      - Endpoint: GET /memorizes
      - Response: List of all memorization records associated with the authenticated user.

2. Get a Specific Memorize
      - Endpoint: GET /memorizes/:id
      - Response: Memorization record with the specified id.

3. Add a Memorize
      - Endpoint: POST /memorizes
      - Request Body

        ``` bash
        {
        "surahName": "Al-Fatiha",
        "ayahRange": "1-7",
        "totalAyah": 7,
        "dateStarted": "2024-09-17T00:00:00Z",
        "dateCompleted": "2024-09-24T00:00:00Z",
        "reviewFrequency": "Weekly",
        "lastReviewDate": "2024-09-17T00:00:00Z",
        "accuracyLevel": "95",
        "nextReviewDate": "2024-09-24T00:00:00Z",
        "notes": "Review after one week"
        }
        ```
    - Response: Status 201 Created with the ID of the new memorization record.

4. Update a Memorize
    - Endpoint: PUT /memorizes/:id
    - Request Body: Similar to the POST /memorizes body, but for updating a specific record.
    - Response: Status 200 OK with the updated record.

5. Delete a Memorize
    - Endpoint: DELETE /memorizes/:id
    - Response: Status 200 OK when the record is successfully deleted.

#### Dummy User Credentials
You can use the following dummy user accounts for testing login:
1. John Doe
    - Username: john_doe
    - Password: password123

2. Jane Doe
    - Username: jane_doe
    - Password: password456

#### JWT Authentication
For any protected route (like accessing or managing memorizes), you will need to include a valid JWT token in the Authorization header.

After logging in via the /signin endpoint, you will receive a token in the response:

``` bash
{
  "status": "Logged in",
  "token": "your-jwt-token-here"
}
```

Include this token in the Authorization header for protected routes:
``` bash
Authorization: Bearer your-jwt-token-here
```

### Database Configuration
Make sure you have a PostgreSQL database running and update the connection settings in main.go:

``` bash
dbCredential := Credential{
    Host:         "localhost",
    Username:     "postgres",
    Password:     "your-password",
    DatabaseName: "your-database",
    Port:         5432,
}
```
### Data Models
- User: Handles user information such as Username and Password.
- Memorize: Tracks Quran memorization progress for a user, including fields like SurahName, AyahRange, TotalAyah, and ReviewFrequency.

### Error Handling
For error responses, the API follows the structure:
``` bash
{
  "error": "Error message here"
}
```

### Further Improvements
- Add proper password hashing for secure storage of passwords.
- Add more robust validation for input fields.
- Implement rate-limiting for authentication requests.

### License
This project is licensed under the MIT License.


### Additional Notes:
- This `README.md` explains how to run the server, the available API endpoints, test execution, and details about the dummy accounts.
- If you add more features in the future, make sure to update the endpoint section accordingly.
