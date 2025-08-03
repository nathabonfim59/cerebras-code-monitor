# Cerebras Code Monitor

Real-time monitoring tool for Cerebras AI usage with rate limit tracking. Track your token consumption and request limits with predictions and warnings.

## Features

- Real-time monitoring of Cerebras API usage
- Rate limit tracking via response headers
- Configurable refresh rates with intelligent display updates
- Advanced terminal UI with color-coded progress bars and tables
- Multi-level warning system when approaching limits
- Automatic timezone detection and 12h/24h time format preferences
- Professional architecture with modular design
- Comprehensive logging capabilities

## Cerebras Rate Limits

Cerebras enforces rate limits per API key with the following headers available in API responses:

| Header | Description |
|--------|-------------|
| `x-ratelimit-limit-requests-day` | Maximum number of requests allowed per day |
| `x-ratelimit-limit-tokens-minute` | Maximum number of tokens allowed per minute |
| `x-ratelimit-remaining-requests-day` | Number of requests remaining for the current day |
| `x-ratelimit-remaining-tokens-minute` | Number of tokens remaining for the current minute |
| `x-ratelimit-reset-requests-day` | Time (in seconds) until daily request limit resets |
| `x-ratelimit-reset-tokens-minute` | Time (in seconds) until per-minute token limit resets |

## Installation

### Prerequisites

- Go 1.24.5 or higher
- Valid Cerebras session token

### Building from Source

```bash
git clone https://github.com/nathabonfim59/cerebras-code-monitor.git
cd cerebras-code-monitor
go build -o cerebras-monitor main.go
```

### Using Go Install

```bash
go install github.com/nathabonfim59/cerebras-code-monitor@latest
```

## Usage

### Authentication

The monitor can authenticate using either a session cookie or an API key.

#### Session Cookie Authentication (Recommended)

This method provides the most accurate data for token prediction calculations:

1. Log into your Cerebras Cloud account at https://cloud.cerebras.ai
2. Extract the session token from browser cookies:
   - Open Developer Tools (F12 or right-click → Inspect)
   - Go to Application tab → Cookies → https://cloud.cerebras.ai
   - Copy the value of 'authjs.session-token' cookie
   ```
   authjs.session-token=your-session-token-here
   ```
3. Set the token as an environment variable:
   ```bash
   export CEREBRAS_SESSION_TOKEN="your-session-token-here"
   ```

Note: The authjs.session-token cookie is HTTP-only, which prevents programmatic access. 
You'll need to manually copy it from your browser's Developer Tools when required. 
This token is only used to fetch your usage data from Cerebras - you can inspect the 
source code yourself as this tool is open source.

#### API Key Authentication (Alternative)

You can also authenticate using a Cerebras API key, though this method has limitations:
- Shows only data for that specific key
- Cannot switch organizations
- Less accurate for token prediction calculations
- Minute-level data is not available
- Each request "wastes" approximately 5 tokens as metadata is extracted from response headers
- Requires longer monitoring intervals to minimize token consumption
- Provides less precise monitoring compared to session token authentication

To use API key authentication:
1. Get your API key from the Cerebras dashboard
2. Set it as an environment variable:
   ```bash
   export CEREBRAS_API_KEY="your-api-key-here"
   ```
3. Or use the login command:
   ```bash
   cerebras-monitor login apikey your-api-key-here
   ```

This will save the API key to your local database at `$XDG_CONFIG_HOME/cerebras-monitor/settings.yaml` (typically `~/.config/cerebras-monitor/settings.yaml`)

### Basic Commands

```bash
# Start monitoring with default settings
cerebras-monitor

# Monitor with custom refresh rate
cerebras-monitor --refresh-rate 5
```

### Configuration Options

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| --session-token | string | "" | Cerebras session token (can be set via environment variable) |
| --org-id | string | "" | Organization ID to monitor |
| --model | string | "qwen-3-coder-480b" | Model to monitor |
| --refresh-rate | int | 10 | Data refresh rate in seconds (1-60) |
| --refresh-per-second | float | 0.75 | Display refresh rate in Hz (0.1-20.0) |
| --timezone | string | auto | Timezone (auto-detected) |
| --time-format | string | auto | Time format: 12h, 24h, or auto |
| --theme | string | auto | Display theme: light, dark, or auto |
| --log-level | string | INFO | Logging level: DEBUG, INFO, WARNING, ERROR, CRITICAL |
| --log-file | path | None | Log file path |
| --debug | flag | false | Enable debug logging |
| --version, -v | flag | false | Show version information |
| --clear | flag | false | Clear saved configuration |

## API Integration

The monitor makes requests to the Cerebras GraphQL endpoint:
`https://cloud.cerebras.ai/api/graphql`

Rate limit information is extracted from response headers:
- Daily request limits and remaining counts
- Per-minute token limits and remaining counts
- Reset times for both limits

## Development

Built with Go using spf13 libraries:
- spf13/cobra for CLI framework
- spf13/viper for configuration management

### Dependencies

- github.com/spf13/cobra
- github.com/spf13/viper
- github.com/cli/browser
- github.com/cli/oauth
- github.com/cli/safeexec
- sqlc (https://sqlc.dev) - required for database query generation

### Building

Before building, ensure you have sqlc installed:

```bash
# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate database code
sqlc generate

# Build the application
go build -o cerebras-monitor main.go
```

### Testing

```bash
go test ./...
```

## License

MIT License

## Contributing

Contributions are welcome. Please fork the repository and submit a pull request with your changes.