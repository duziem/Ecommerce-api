## API Documentation
https://www.postman.com/cloudy-sunset-241894/new-workspace/documentation/755tanx/e-commerce-api

## App requirements
* Need to install GO [Go installation](https://go.dev/doc/install)
* Need to install GOlang-migration cli to be able to run migrations [GOlang-migration CLI](https://github.com/golang-migrate/migrate/tree/v4.17.0/cmd/migrate)
* Need to install Postgres. The app uses a Postgres Database [Postgres download](https://www.postgresql.org/download/)

## Step by Step
* Clone the repo ```git clone <repo path>```
* Create a database called ecommerce_db. This can be done using psql by running the SQL statement ```CREATE database ecommerce_db```
* Run migrations
  * Steps:
    * Run these command to create tables
    ```bash
      Make migrate-up
    ```
    or alternatively
    ```bash
        go run cmd/migrate/main.go up
    ```
    * Run these command to delete tables
    ```bash
      Make migrate-down
    ```
    or alternatively
    ```bash
      go run cmd/migrate/main.go down
    ```
* Navigate to the repo -> run the application using the command: ```go run cmd/main.go```
* Test the application by sending requests using tools like Postman, swagger, etc.
* Some endpoints are restricted to admins. By default a user is created with a role "user" but can be given Admin permissions by updating the user's role attribute. The easiest way to do this is by performing the following steps
  * Call the register-user endpoint in the documentation
  * Next, call the login-user endpoint to login and create a token. The token is automatically assigned to an environment variable to enable authenticated requests pass
  * Next, call the get-user endpoint to get the user id
  * Lastly, call the update-user endpoint to update your user's role to admin
* This can also be done via psql by running the sql statement ```UPDATE users set role = 'admin' where id = $userID``` and $userID should be replaced with the actual user ID

## File structure
* Cmd
### Sub dirs:
* Api
  * Cmd/Api/api.go - contains functions for creating a new Api server and running the Api server
  * Cmd/migrate/migrations - contains the migration files
* migrate
  * Cmd/migrate/main.go - contains the script for running migrations
* Cmd/main.go - This is the application entry point

* Db
  * Db/db.go - contains the database config

* Types
  * Types/types.go - contains object schema

* Utils
  * Utils/utils.go - contains helper functions

* Services
### Sub dirs
* Auth:
  * Services/auth/jwt.go - contains functions for creating and validating the JWT
  * Services/auth/password.go - Contains functions for password having

* User
  * User/routes.go
  * User/store.go

* Product
  * Product/routes.go
  * Product/store.go

* Order
  * Order/routes.go
  * Order/store.go

* Cart
  * Cart/routes.go

