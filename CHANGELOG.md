# Harpy Changelog

## [1.0.0] - Initial Release - 2025-09-27

This is the first release of **Harpy** - A high-performance web reconnaissance tool for extracting endpoints, parameters, and hidden assets from JavaScript files.

### Core Features
- **Advanced JavaScript Analysis**: Hybrid regex and AST-based parsing for accurate endpoint and parameter extraction from JavaScript files.
- **Dual Analysis Modes**:
  - **Regex Mode** (`--enable-regex`): Ultra-fast pattern matching for quick reconnaissance
  - **AST Mode** (`--enable-ast`): Deep JavaScript parsing for comprehensive analysis with higher accuracy
  - **Hybrid Mode**: Intelligent combination of both approaches for optimal results
- **Intelligent Extraction**: Discovers:
  - API endpoints and paths
  - URL parameters and form fields  
  - HTTP headers and authentication tokens
  - Internal domains and subdomains
  - Hidden admin panels and debug endpoints
- **Concurrent Architecture**: High-speed processing with configurable worker pools (`-c` flag)
- **Adaptive Rate Limiting**: Smart per-domain throttling (`-l` flag) that automatically adjusts to server responses
- **Flexible Input Sources**: URLs (`-u`), files (`-f`), directories (`-d`), or stdin for seamless integration
- **Professional Output**: Clean terminal output plus structured JSON (`--json`) for automation
- **Advanced Configuration**:
  - File size limits (`--max-file-size`, `--no-limit`)
  - Custom headers (`-H`)
  - Proxy support (`-p`, `--proxy`)
  - TLS options (`--skip-verify`)
  - Output control (`-v`, `--silent`, `--no-color`) 