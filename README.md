# grocery-planner
Plan groceries by scheduling meals for a period of time

## Development Setup
These programs should be installed:
1. Go 1.22.5
2. Docker Engine
3. Make
4. [Golang Migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

### Setup database in Docker Engine
1. Pull postgres image by using this command:

        docker pull postgres
2. Create .env file according to env.example
    1. **POSTGRES_USER**: Username for the user that has the permission to read and write the database
    2. **POSTGRES_PASSWORD**: Password for the database user
    3. **POSTGRES_HOST**: Domain address for the database host server
    4. **SERVER_ADDRESS**: Domain and port address for the API server
    5. **SYM_KEY**= Secret key for authorization
    6. **ACCESS_TOKEN_DURATION**: Authorization token duration in minutes
        
3. Run these make commands from the project directory in order:
        
        make postgres
        make createuser
        make createdb
4. If you previously had setup a postgres database, run this command to start the container:
        
        make runpostgres
5. Migrate the database scheme:

        make migrateup
        // for migrate down use 'make migratedowm'
5. To stop the running postgres container:

        make stoppostgres


### Setup Go Modules
Use this command to install the missing modules and remove the unused modules:

        go mod tidy

### Run Go Server
Use this command to start the server:

        make server
        