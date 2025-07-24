# vxFormsUI

GO service for processing ingest related forms
This project is a Go module that uses the [Gin web framework](https://gin-gonic.com/) and [Bootstrap 5](https://getbootstrap.com/) to implement a web-based UI with the following features:

- **Form Selection:** Users can choose which form to use from a main page.
- **Dynamic Form Rendering:** The UI presents the appropriate form for creating the associated JSON document based on user input.
- **Default Version Value:** Any input field named `version` is pre-filled with the default value `"V01"`.
- **Back Navigation:** Each form includes a back button that returns the user to the main page.

## Running with Docker and Docker Compose

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) installed
- A credentials file named `CREDENTIALS_FILE` in the user home directory (or update the path in `docker-compose.yml`)

### Build and Run

1. Place your credentials in a file named `CREDENTIALS_FILE` in the home directory.
2. Build and start the container using Docker Compose from the `docker` directory:

   ```sh
   cd docker
   docker compose up --build
   ```

3. The app will be available at [http://localhost:8080](http://localhost:8080)

### Stopping

To stop the app, press `Ctrl+C` in the terminal and run:

```sh
docker compose down
```

### Notes

- The `CREDENTIALS_FILE` is mounted as a Docker secret and available in the container at `/run/secrets/CREDENTIALS_FILE`.
- You can change the port mapping in `docker-compose.yml` if needed.
