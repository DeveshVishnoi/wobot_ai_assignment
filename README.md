# WOBOT AI Assignment

## Prerequisites

- **Go** – [Download Go](https://golang.org/dl/)
- **MongoDB** – [Download MongoDB](https://www.mongodb.com/try/download/)

## Run Locally

### Clone the project

```bash
git clone https://github.com/DeveshVishno/Wobot_AI_Assignment.git
```

Go to the project directory

```bash
  cd Wobot_AI_Assignment
```

To run the project locally

```bash
  go build
  ./file_upload.exe
```

### End Points

-`/login` -- User login

`/register` -- Create a new user

`/storage/remaining` -- Get remaining storage for the logged-in user

`/upload` -- Upload a file

`/files` -- Get all uploaded files for the user

### Cofiguration file available on this location (env)

```bash
config/config.json
```

Note: Currently, MongoDB is configured to run on 127.0.0.1 (localhost).
You can change this in the config/config.json file according to your MongoDB setup.
