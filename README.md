# bindle

> [!WARNING] This is a personal learning project. While the utilities work and are tested, I wouldn't recommend using this in production just yet. I'm trying to get better at Go and these APIs might change as I figure out better patterns. Use at your own risk, or better yet, grab the code and adapt it to your needs.

A collection of utilities for common problems you'll run into building Go applications.

The idea is simple: just grab a helper function for a common problem and move on...

## Installation

```bash
go get github.com/ddddami/bindle
```

## What's in the bindle

### `jsonx`

JSON response boilerplate is definitely one of the most tiring things to deal with when building APIs. This helps standardize responses across your app.

```go
import "github.com/ddddami/bindle/jsonx"
```

#### Basic responses

```go
func getUsers(w http.ResponseWriter, r *http.Request) {
  users := []User{{ID: 1, Name: "Dami", Email: "dami@dami.dev"}, {ID: 2, Name: "Ngozi A.", Email: "nga@gm.com"}}

    // Simple success response
    jsonx.Send(w, users)
}
```

**Response:**

```json
[
  {
    "id": 1,
    "name": "Dami",
    "email": "dami@dami.dev"
  },
  {
    "id": 2,
    "name": "Ngozi A.",
    "email": "nga@gm.com"
  }
]
```

#### Creating resources

```go
func createUser(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }

    // Decode JSON from request
    if err := jsonx.DecodeJSONFromRequest(r, &input); err != nil {
        jsonx.SendError(w, err)
        return
    }

    newUser := User{ID: 123, Name: input.Name, Email: input.Email}

    // Custom status with headers
    jsonx.RespondWithJSON(w, newUser, jsonx.Options{
        SuccessStatus: http.StatusCreated,
        Headers: map[string]string{
            "X-Resource-ID": "123",
        },
    })
}
```

**Response:**

```json
{
  "id": 123,
  "name": "Dami",
  "email": "dami@dami.dev"
}
```

#### Error handling

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Simple error
    jsonx.SendError(w, errors.New("something went wrong"))
}
```

**Response:**

```json
{
  "error": "something went wrong"
}
```

#### Custom errors with details

```go
func handler(w http.ResponseWriter, r *http.Request) {
    customErr := jsonx.ErrorDetail{
        Code:    "RESOURCE_NOT_FOUND",
        Message: "The requested resource could not be found",
    }

    jsonx.RespondWithError(w, customErr, jsonx.Options{
        ErrorStatus:    http.StatusNotFound,
        IndentResponse: true,
    })
}
```

**Response:**

```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "The requested resource could not be found"
  }
}
```

#### Complex responses with metadata

```go
func handler(w http.ResponseWriter, r *http.Request) {
    userList := UserList{Users: users, Total: len(users)}
    meta := Pagination{CurrentPage: 1, TotalPages: 5, PerPage: 10}

    jsonx.RespondWithSuccess(w, userList, meta, jsonx.Options{
        IndentResponse: true,
    })
}
```

**Response:**

```json
{
    "data": {
        "users": [...],
        "total": 25
    },
    "meta": {
        "current_page": 1,
        "total_pages": 5,
        "per_page": 10
    }
}
```

### `uploads`

Handles file uploads and downloads

```go
import "github.com/ddddami/bindle/uploads"
```

#### Single file upload

```go
func handleUpload(w http.ResponseWriter, r *http.Request) {
    opts := uploads.FileUploadOptions{
        DestinationDir:    "./uploads",
        MaxSize:           5 * 1024 * 1024, // 5MB
        AllowedExts:       []string{"jpg", "jpeg", "png", "pdf"},
        RandomizeFilename: true,
        FilenamePrefix:    "upload_",
    }

    savedFile, err := uploads.SaveSingleFormFile(r, "file", &opts)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    fmt.Printf("Uploaded: %s (saved as %s)", savedFile.OriginalName, savedFile.SavedName)
}
```

#### Multiple file upload

```go
func handleMultipleUpload(w http.ResponseWriter, r *http.Request) {
    opts := uploads.FileUploadOptions{
        DestinationDir:    "./uploads",
        MaxSize:           10 * 1024 * 1024,
        AllowedExts:       []string{"jpg", "png", "pdf"},
        RandomizeFilename: true,
    }

    savedFiles, err := uploads.SaveMultipleFormFiles(r, "files", &opts)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    fmt.Printf("Uploaded %d files", len(savedFiles))
}
```

#### File downloads

```go
func handleDownload(w http.ResponseWriter, r *http.Request) {
    filePath := "./uploads/document.pdf"

    opts := uploads.DownloadOptions{
        ForceDownload:     true,
        SuggestedFilename: "my-document.pdf",
        ExtraHeaders: map[string]string{
            "X-Download-Type": "Document",
        },
    }

    if err := uploads.ServeFileForDownload(w, r, filePath, opts); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
```

### `fs`

File system operations.

```go
import "github.com/ddddami/bindle/fs"
```

#### Directory creation

```go
func main() {
    // Creates nested directories if they don't exist
    err := fs.CreateDirIfNotExistsWithPerm("./data/uploads/2024", 0o755)
    if err != nil {
        log.Fatal(err)
    }
}
```

#### File operations

```go
func main() {
    if fs.PathExists("./config.json") {
        fmt.Println("Config file found")
    }

    err := fs.CopyFile("./source.txt", "./backup/source.txt")
    if err != nil {
        log.Fatal(err)
    }

    txtFiles, err := fs.ListFilesWithExt("./documents", ".txt")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d .txt files\n", len(txtFiles))

    // Safe removal (won't panic if path doesn't exist)
    fs.SafeRemove("./temp/old-data")
}
```

### `strutil`

Utility around string manipulation.

```go
import "github.com/ddddami/bindle/strutil"
```

#### URL sanitization

```go
func main() {
    title := "This is a Test Article!--- With Spaces & Special Characters?"
    slug := strutil.SanitizeURL(title)
    // Output: "this-is-a-test-article-with-spaces-special-characters"
}
```

#### Text related stuff

```go
func main() {
    longText := "This is a very long text that needs to be truncated..."
    short := strutil.Truncate(longText, 30)
    // Output: "This is a very long text th..."

    html := "<p>This is <strong>bold</strong> text with <a href='#'>a link</a>.</p>"
    plain := strutil.StripHTML(html)
    // Output: "This is bold text with a link."

    // Clean filenames
    badName := "User's Document: Important! (2023).txt"
    safeName := strutil.FormatFilename(badName)
    // Output: "Users_Document_Important_2023.txt"
}
```

#### Email validation

```go
func main() {
    emails := []string{"user@example.com", "not-an-email", "missing@domain"}

    for _, email := range emails {
        if strutil.IsValidEmail(email) {
            fmt.Printf("%s is valid\n", email)
        } else {
            fmt.Printf("%s is invalid\n", email)
        }
    }
}
```

### `random`

Generate random strings when you need them.

```go
import "github.com/ddddami/bindle/random"
```

```go
func main() {
    randomStr, err := random.Generate(random.Options{Length: 10})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Random string:", randomStr)
}
```

## Demo app

Check out [bindle-app](https://github.com/ddddami/bindle-app) for a working example showing all these utilities in action and see more usage examples.

## Why?

Basically because I want to get better at writing Go and since I intend on writing a lot of it in the future, I decided to write (and refactor along the way) a bunch of helpers for repetitive and common patterns in web dev.

## Stuff I might add in the future

- `config` - Configuration manager
- `log` - A tiny structured logger
- `validator` - Request validation

## Contributing

If you want to make any change or a suggestion, you can open an issue or submit a PR.
