module github.com/LukeHagar/plexgo/integration_tests

go 1.22

require (
	github.com/LukeHagar/plexgo v0.21.2
	github.com/joho/godotenv v1.5.1
)

require github.com/ericlagergren/decimal v0.0.0-20221120152707-495c53812d05 // indirect

// Use the local version of plexgo for development
replace github.com/LukeHagar/plexgo => ../
