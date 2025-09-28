# Harpy

![Go Version](https://img.shields.io/github/go-mod/go-version/rafabd1/Harpy)
![Release](https://img.shields.io/github/v/release/rafabd1/Harpy?include_prereleases)
![Build Status](https://github.com/rafabd1/Harpy/workflows/Release%20Harpy/badge.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![GitHub stars](https://img.shields.io/github/stars/rafabd1/Harpy?style=social)
![Go Report Card](https://goreportcard.com/badge/github.com/rafabd1/Harpy)

<div align="center">

<pre>
    ██╗  ██╗ █████╗ ██████╗ ██████╗ ██╗   ██╗
    ██║  ██║██╔══██╗██╔══██╗██╔══██╗╚██╗ ██╔╝
    ███████║███████║██████╔╝██████╔╝ ╚████╔╝ 
    ██╔══██║██╔══██║██╔══██╗██╔═══╝   ╚██╔╝  
    ██║  ██║██║  ██║██║  ██║██║        ██║   
    ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝        ╚═╝   
</pre>

</div>

<p align="center">
    <b>High-performance web reconnaissance tool for extracting endpoints, parameters, and hidden assets from Web files.</b>
</p>

## ✨ Features

- **🧠 Intelligent Hybrid Analysis**: Automatically combines regex and AST-based parsing for maximum accuracy with smart fallbacks
- **🚀 High Performance**: Multi-threaded processing with configurable worker pools for blazing-fast analysis
- **🎯 Comprehensive Extraction**: Discovers API endpoints, URL parameters, HTTP headers, domains, and hidden admin panels
- **⚡ Adaptive Rate Limiting**: Smart per-domain throttling that automatically adjusts based on server responses
- **📁 Flexible Input**: URLs, local files, directories, or stdin for seamless tool integration
- **🎨 Professional Output**: Clean terminal display plus structured JSON for automation and CI/CD pipelines
- **🔒 Intelligent Filtering**: Advanced pattern recognition to minimize false positives while maintaining high accuracy

### What Harpy Extracts

- **API Endpoints**: REST endpoints, GraphQL schemas, internal APIs
- **Parameters**: URL parameters, form fields, API parameters, path parameters
- **Headers**: Authorization tokens, API keys, custom headers, CORS configurations  
- **Domains**: Internal domains, subdomains, CDN endpoints, third-party services
- **Hidden Assets**: Admin panels, debug endpoints, development URLs

## 📦 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/rafabd1/Harpy.git
cd Harpy

# Build the binary
go build -o harpy ./cmd/harpy

# Optional: Move to path (Linux/macOS)
sudo mv harpy /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/rafabd1/Harpy/cmd/harpy@latest
```

### Binary Releases

Download pre-built binaries for your platform from the [releases page](https://github.com/rafabd1/Harpy/releases).

## 🚀 Quick Start

Extract endpoints from a single JavaScript file:

```bash
harpy -u https://example.com/assets/app.js
```

Scan local JavaScript files:

```bash
harpy -d /path/to/js/files
```

Scan from a list of targets and output as JSON:

```bash
harpy -f targets.txt --json -o results.json
```

Extract with verbose output:

```bash
harpy -u https://target.com/main.js -v
```

## 🔧 Command Line Options

### Input Options

| Flag | Description | Default |
|------|-------------|---------|
| `-u` | Single target URL or file path | - |
| `-f` | File containing list of targets (one per line) | - |
| `-d` | Directory to scan for JavaScript/TypeScript/HTML/JSON files | - |

### Analysis Options

| Flag | Description | Default |
|------|-------------|---------|
| `-c` | Number of concurrent workers | `25` |
| `--max-file-size` | Maximum file size to process (KB) | `10240` |
| `--no-limit` | Disable file size restrictions | `false` |

**Note:** Harpy automatically uses intelligent hybrid analysis (Regex + AST) with smart fallbacks. No manual mode configuration needed.

### Network Options

| Flag | Description | Default |
|------|-------------|---------|
| `-l` | Requests per second limit per domain | `30` |
| `-t` | Request timeout in seconds | `10` |
| `-H` | Custom headers (can be used multiple times) | - |
| `-p` | Proxy list file | - |
| `--proxy` | Single proxy server | - |
| `--skip-verify` | Skip TLS certificate verification | `false` |

### Output Options

| Flag | Description | Default |
|------|-------------|---------|
| `-o` | Output file path | stdout |
| `--json` | Output results in JSON format | `false` |
| `-v` | Enable verbose output | `false` |
| `--silent` | Suppress all output except findings | `false` |
| `--no-color` | Disable colored output | `false` |

## 💡 Example Usage

### Bug Bounty Reconnaissance

```bash
# Quick scan for endpoints and parameters
harpy -u https://target.com/app.js

# Comprehensive directory scan with JSON output
harpy -d ./js-files --json -o findings.json

# Scan with custom headers and proxy
harpy -f targets.txt -H "User-Agent: Harpy/1.0" --proxy http://127.0.0.1:8080
```

### Integration with Other Tools

```bash
# Pipe from subfinder and httpx
subfinder -d target.com | httpx -path /assets/app.js | harpy

# Extract endpoints and pass to ffuf
harpy -u https://target.com/main.js --json | jq -r '.results.findings[].endpoints[].path' | ffuf -u https://target.com/FUZZ -w -
```

## 📊 Output Formats

### Terminal Output

```
=== Extraction Results ===
✓ Sources processed: 1
✓ Domains found: 3  
✓ Endpoints found: 12
✓ Headers found: 4
```

### JSON Output

```json
{
  "metadata": {
    "tool": "Harpy",
    "version": "1.0.0",
    "timestamp": "2024-01-15T10:30:00Z",
    "total_sources": 1
  },
  "results": {
    "findings": [
      {
        "source": "https://example.com/app.js",
        "domains": ["api.example.com", "admin.example.com"],
        "endpoints": [
          {
            "method": "GET",
            "path": "/api/v1/users",
            "context": "fetch('/api/v1/users')"
          }
        ],
        "parameters": [
          {
            "name": "userId",
            "type": "url",
            "context": "?userId=123"
          }
        ],
        "headers": [
          {
            "name": "Authorization",
            "context": "headers: {'Authorization': token}"
          }
        ]
      }
    ]
  },
  "summary": {
    "total_domains": 2,
    "total_endpoints": 1,
    "total_parameters": 1,
    "total_headers": 1
  }
}
```

## ⚠️ Disclaimer

**Usage Warning & Responsibility**

This tool is designed for security professionals, bug bounty hunters, and researchers for legitimate testing purposes only. Users must have explicit permission to test any target. The author is not responsible for any misuse or damage caused by this program.

## 📚 Documentation

- [Architecture Overview](dev/HARPY-ARCHITECTURE.md) - Technical details and design decisions
- [Changelog](CHANGELOG.md) - Version history and updates

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<p align="center">
    <sub>Made with 🔍 for bug bounty hunters and security researchers</sub>
</p>

<p align="center">
    <sub>Created by Rafael (github.com/rafabd1)</sub>
</p>



### Bug Bounty Reconnaissance    <a href="https://ko-fi.com/rafabd1" target="_blank"><img src="https://storage.ko-fi.com/cdn/kofi2.png?v=3" alt="Buy Me A Coffee" style="height: 60px !important;"></a>

```bash</p> -->

# Quick scan for endpoints and parameters
harpy -u https://target.com/app.js --enable-regex --enable-ast

# Comprehensive directory scan with JSON output
harpy -d ./js-files --enable-ast --json -o findings.json

# Scan with custom headers and proxy
harpy -f targets.txt -H "User-Agent: Harpy/1.0" --proxy http://127.0.0.1:8080
```

