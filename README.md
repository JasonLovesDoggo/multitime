# MultiTime

MultiTime is a WakaTime proxy that forwards your coding activity to multiple WakaTime-compatible backends simultaneously. Perfect for those who want to track their coding activity across multiple platforms or maintain a backup of their WakaTime data.

## Features

- Forward WakaTime heartbeats to multiple backends
- Designate a primary backend for status queries
- Support for both official WakaTime and compatible backends (like [Hack Club HighSeas](https://highseas.hackclub.com/))
- Minimal configuration required
- Debug logging for troubleshooting

## Installation
(Requires golang 1.23+)
```bash
# Clone the repository
go install github.com/JasonLovesDoggo/multitime@latest
```

## Configuration

Create a `config.toml` file:

```toml
port = 3005 # can be any port you want
debug = true # Optional, enables debug logging
log_file = "multitime.log"  # Optional, logs to stdout if not specified

[[backends]]
name = "Official WakaTime"
url = "https://wakatime.com"
api_key = "your-wakatime-api-key"
is_primary = true  # Primary backend for status queries

[[backends]]
name = "Hack Club HighSeas"
url = "https://waka.hackclub.com"
api_key = "your-highseas-api-key"
is_primary = false

# Add more backends as needed
```

### Backend Configuration

- `name`: Identifier for the backend (used in logs)
- `url`: Base URL of the WakaTime-compatible API
- `api_key`: Your API key for that backend
- `is_primary`: Set to `true` for one backend only - used for status queries

## Usage

1. Start the server:
```bash
multitime config.toml
```

2. Configure your WakaTime client:
   - Find your IDE's WakaTime plugin settings
   - Set the API URL to `http://localhost:3000` (if you don't see a setting, try editing `~/.wakatime.cfg`)
   - Set any valid string as the API key (it will be replaced with the correct key for each backend)

### Using with Hack Club HighSeas

[Hack Club HighSeas](https://highseas.hackclub.com/) is a self-hosted WakaTime-compatible backend. To use MultiTime with HighSeas:

1. Set up your HighSeas instance
2. Get your API key from HighSeas
3. Add HighSeas as a backend in your `config.toml`:
```toml
[[backends]]
name = "HackClub WakaTime"
url = "https://waka.hackclub.com""
api_key = "your-highseas-api-key"
is_primary = false  # true if you want to use HighSeas for status queries
```

## Supported Endpoints

MultiTime currently supports these WakaTime API endpoints:

### POST `/users/current/heartbeats
- Forwards coding activity heartbeats to all configured backends
- Returns the response from the primary backend
- Adds custom user agent identifier

### POST `/users/current/heartbeats.bulk`
- Forwards multiple heartbeats to all configured backends
- Returns the response from the primary backend
- Adds custom user agent identifier

### GET `/users/current/statusbar/today`
- Retrieves today's coding activity summary from the primary backend
- Used by IDE plugins for status bar updates
- Returns cached data if available, empty summary if not

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)

## Credits

Created by [JasonLovesDoggo](https://github.com/JasonLovesDoggo)

Special thanks to:
- [WakaTime](https://wakatime.com) for their amazing time tracking platform
- [Hack Club](https://hackclub.com) for creating HighSeas
