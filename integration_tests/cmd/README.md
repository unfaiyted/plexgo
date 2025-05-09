# Helper Scripts for Integration Testing

This directory contains helper scripts for setting up and running integration tests.

## Finding Media IDs

To find media IDs for testing, run:

```bash
cd cmd
go run find_media_ids.go
```

This script will:
1. Load your `.env` file from the project root
2. Connect to your Plex server
3. List the first 20 media items from the section specified in `PLEX_SECTION_ID`
4. Show example IDs to use in your `.env` file

### Prerequisites

Make sure you have set up the `.env` file in the project root directory with:

```
PLEX_SERVER_PROTOCOL=https
PLEX_SERVER_IP=your_plex_server_ip
PLEX_SERVER_PORT=32400
PLEX_TOKEN=your_plex_token
PLEX_SECTION_ID=your_library_section_id
```

### Output Example

```
Getting items from library section 1...
Found 157 items in section 1:

RatingKey | Title
----------|------------------
12345 | The Shawshank Redemption
12346 | The Godfather
12347 | The Dark Knight
12348 | Pulp Fiction
12349 | Fight Club
...and more. Showing only first 20 items.

Example for .env file:
PLEX_TEST_MEDIA_IDS=12345,12346,12347
```

Copy the example line to your `.env` file to use these media IDs in your integration tests.