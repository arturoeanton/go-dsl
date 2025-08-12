# HTTP DSL Project Structure

## 📁 Project Organization

```
http_dsl/
├── universal/              # Core DSL implementation
│   ├── http_dsl.go        # DSL grammar and parser
│   ├── http_engine.go     # HTTP client engine with advanced features
│   └── http_dsl_test.go   # Comprehensive unit tests
│
├── runner/                 # Script execution engine
│   └── http_runner.go     # .http file parser and executor
│
├── scripts/               # Demo scripts with real APIs
│   ├── quick_demo.http    # 30-second quick demonstration
│   ├── demo_simple.http   # Complete feature showcase
│   ├── demo_github.http   # GitHub API integration
│   ├── demo_weather.http  # Weather API with logic
│   ├── demo_basic.http    # Basic operations test
│   └── demo_advanced.http # Complex workflow simulation
│
├── main.go                # Example usage program
├── README.md              # English documentation
├── README.es.md           # Spanish documentation
├── PROJECT.md             # This file
├── go.mod                 # Go module definition
└── go.sum                 # Dependency checksums
```

## 🚀 Quick Setup

```bash
# 1. Build the runner
go build -o http-runner ./runner/http_runner.go

# 2. Run a demo
./http-runner scripts/quick_demo.http

# 3. Create your own script
echo 'GET "https://api.github.com"' > my_test.http
./http-runner my_test.http
```

## 🎯 Key Components

### Universal Package
The core DSL implementation following the universal pattern:
- Self-contained and portable
- No external dependencies beyond go-dsl
- Enterprise-grade HTTP client features

### Runner
Command-line tool for executing .http scripts:
- Real-time execution feedback
- Performance metrics
- Error handling and debugging

### Scripts
Production-ready demo scripts using real APIs:
- No API keys required
- Public APIs only
- Comprehensive feature coverage

## 📊 Test Results

✅ **30+ successful API calls** to real services
✅ **All HTTP methods** tested and working
✅ **JSONPath extraction** functioning correctly
✅ **Variables and conditionals** operational
✅ **Performance assertions** validated

## 🔧 Development

```bash
# Run tests
go test ./universal -v

# Run with debugging
./http-runner -v scripts/demo_simple.http

# Check coverage
go test -cover ./universal

# Build for distribution
go build -ldflags="-s -w" -o http-runner ./runner/http_runner.go
```

## 📝 Creating Scripts

Basic script structure:
```http
# Set variables
set $base_url "https://api.example.com"

# Make request
GET "$base_url/users"
assert status 200

# Extract data
extract jsonpath "$.data[0].id" as $user_id

# Use in next request
GET "$base_url/users/$user_id"
```

## 🌟 Features

- ✅ All HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, CONNECT, TRACE)
- ✅ Variables and data extraction
- ✅ Conditionals and loops
- ✅ Response assertions
- ✅ Cookie management
- ✅ SSL/TLS configuration
- ✅ Proxy support
- ✅ Rate limiting
- ✅ Retry policies
- ✅ OAuth 2.0
- ✅ GraphQL queries
- ✅ Session management
- ✅ Performance metrics

## 📚 Documentation

- [English README](README.md) - Full documentation in English
- [Spanish README](README.es.md) - Documentación completa en español
- [Demo Scripts](scripts/) - Working examples with real APIs

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch
3. Add tests for new features
4. Ensure all tests pass
5. Submit a pull request

## 📄 License

Part of the go-dsl project. See parent repository for license details.