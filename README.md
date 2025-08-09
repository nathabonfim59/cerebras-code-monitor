# Cerebras Code Monitor

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/nathabonfim59/cerebras-code-monitor)](https://github.com/nathabonfim59/cerebras-code-monitor/releases)

**Never run out of tokens again!** Monitor your Cerebras AI usage in real-time with rate limit tracking, usage predictions, and warnings before you hit your limits.

<img width="543" height="1063" alt="image" src="https://github.com/user-attachments/assets/69fe1072-478b-48e0-86b4-a8ab2f71f1f4" />

## Quick Start

**1. Install** (choose one):
```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/nathabonfim59/cerebras-code-monitor/main/install.sh | bash

# Windows (PowerShell)
iwr -useb https://raw.githubusercontent.com/nathabonfim59/cerebras-code-monitor/main/install.ps1 | iex
```

**2. Get your session token:**
- Go to [cloud.cerebras.ai](https://cloud.cerebras.ai) and sign in
- Press F12 → Application → Cookies → copy the `authjs.session-token` value
- Set it: `export CEREBRAS_SESSION_TOKEN="your-token-here"`
- Or save it permanently in `~/.config/cerebras-monitor/settings.yaml` (Windows: `%APPDATA%\cerebras-monitor\settings.yaml`)

**3. Start monitoring:**
```bash
cerebras-monitor
```

That's it!

## What You Get

- **Real-time dashboard** - See your usage update live
- **Rate limit tracking** - Never hit unexpected limits
- **Multi-organization support** - Switch between orgs easily
- **Usage predictions** - Know when you'll hit your limits
- **Token consumption monitoring** - Track every request
- **Clean terminal interface** - Beautiful, responsive display

<img width="1909" height="641" alt="image" src="https://github.com/user-attachments/assets/a760f826-daec-4c67-bc02-fca9e7f1d6ab" />


## Coming Soon
- Automatic request interception
- Smart alerts and warnings
- Historical usage trends
- Export capabilities

## Basic Usage

```bash
# Start with default settings
cerebras-monitor

# Custom refresh rate
cerebras-monitor --refresh-rate 5

# Choose organization
cerebras-monitor --org-id your-org-id
```

<details>
<summary>More Installation Options</summary>

### Manual Download
Download from the [releases page](https://github.com/nathabonfim59/cerebras-code-monitor/releases).

### Using Go Install
```bash
go install github.com/nathabonfim59/cerebras-code-monitor/cmd@latest
```

### Building from Source
```bash
git clone https://github.com/nathabonfim59/cerebras-code-monitor.git
cd cerebras-code-monitor
go build -o cerebras-monitor cmd/main.go
```

</details>

<details>
<summary>Authentication Details</summary>

### Session Cookie Authentication (Recommended)
Provides the most accurate data and full organization access.

1. Log into [Cerebras Cloud](https://cloud.cerebras.ai)
2. Extract session token from browser cookies:
   - Open Developer Tools (F12)
   - Go to Application → Cookies → https://cloud.cerebras.ai
   - Copy the `authjs.session-token` value
3. Set as environment variable or save in config file

**Note:** The session token is HTTP-only and must be manually copied. This tool only uses it to fetch your usage data - source code is available for inspection.

### API Key Authentication (Alternative)
Limited functionality compared to session token:
- Shows only data for that specific key
- Cannot switch organizations  
- Less accurate predictions
- Each request consumes ~5 tokens for metadata

To use:
```bash
export CEREBRAS_API_KEY="your-api-key"
# or
cerebras-monitor login apikey your-api-key
```

</details>

<details>
<summary>Configuration Options</summary>

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| --session-token | string | "" | Cerebras session token |
| --org-id | string | "" | Organization ID to monitor |
| --model | string | "qwen-3-coder-480b" | Model to monitor |
| --refresh-rate | int | 10 | Data refresh rate in seconds (1-60) |
| --refresh-per-second | float | 0.75 | Display refresh rate in Hz (0.1-20.0) |
| --timezone | string | auto | Timezone (auto-detected) |
| --time-format | string | auto | Time format: 12h, 24h, or auto |
| --theme | string | auto | Display theme: light, dark, or auto |
| --log-level | string | INFO | Logging level |
| --icons | string | emoji | Icon set: emoji or nerdfont |

</details>

<details>
<summary>Understanding Cerebras Rate Limits</summary>

Cerebras enforces rate limits per API key with these response headers:

| Header | Description |
|--------|-------------|
| `x-ratelimit-limit-requests-day` | Maximum requests per day |
| `x-ratelimit-limit-tokens-minute` | Maximum tokens per minute |
| `x-ratelimit-remaining-requests-day` | Requests remaining today |
| `x-ratelimit-remaining-tokens-minute` | Tokens remaining this minute |
| `x-ratelimit-reset-requests-day` | Daily limit reset time (seconds) |
| `x-ratelimit-reset-tokens-minute` | Minute limit reset time (seconds) |

</details>

<details>
<summary>Development & Contributing</summary>

### Built With
- Go with spf13/cobra for CLI
- spf13/viper for configuration
- sqlc for database queries

### Prerequisites
- Go 1.24.5 or higher
- sqlc (for database code generation)

### Building
```bash
# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate database code
sqlc generate

# Build
go build -o cerebras-monitor main.go
```

### Testing
```bash
go test ./...
```

### Release Process
```bash
# Create and push tag
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0

# Test locally
make release-dry

# Create snapshot
make snapshot
```

### API Integration
Makes requests to: `https://cloud.cerebras.ai/api/graphql`

Rate limit data extracted from response headers.

</details>

## License

MIT License

## Contributing

Contributions welcome! Fork the repository and submit a pull request.
