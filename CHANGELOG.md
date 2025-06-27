# DepScout Changelog

## [0.1.0] - Initial Release

This is the first public release of **DepScout**.

### Core Features
- **Dependency Confusion Scanning**: Identifies potential Dependency Confusion vulnerabilities by scanning JavaScript files for package names and checking their existence on the public npm registry.
- **Dual Parsing Engine**:
  - **Regex Mode**: Fast, high-performance scanning using fine-tuned regular expressions to find `require()` and `import` statements.
  - **Deep Scan Mode (`--deep-scan`)**: High-accuracy scanning using a full Abstract Syntax Tree (AST) parser to eliminate false positives from non-code contexts (e.g., comments). Includes an automatic fallback to regex if AST parsing fails.
- **Concurrent Architecture**: Leverages a powerful, concurrent worker model to perform high-speed fetching and analysis of hundreds or thousands of files simultaneously.
- **Adaptive Rate Limiting**: Features a smart, per-domain rate limiter that automatically adjusts request speeds based on server responses (e.g., `429 Too Many Requests`), maximizing speed without overwhelming targets.
- **Flexible Input**: Accepts targets from single URLs (`-u`), files (`-f`), local directories (`-d`), or piped via `stdin`.
- **Advanced Configuration**: Provides a rich set of CLI flags for fine-grained control:
    - Concurrency levels (`-c`)
    - Request timeouts (`-t`) and custom headers (`-H`)
    - File size limits (`--max-file-size`)
    - Proxy support (`-p`)
- **Multiple Output Formats**: Delivers results in human-readable text or machine-readable `JSON` (`--json`), suitable for direct analysis or integration into other tools. 