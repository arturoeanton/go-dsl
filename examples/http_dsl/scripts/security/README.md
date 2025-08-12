# HTTP DSL v3 - Security Testing Suite 🛡️

A comprehensive collection of defensive security testing scripts for HTTP APIs using HTTP DSL v3. These demos focus exclusively on **defensive security testing** - validating security controls, testing proper authentication, and ensuring APIs follow security best practices.

## 🎯 Security Focus: Defensive Testing Only

**⚠️ IMPORTANT: These demos are designed for DEFENSIVE security testing only:**
- ✅ **Validating security headers and controls**
- ✅ **Testing authentication and authorization**  
- ✅ **Verifying input validation and sanitization**
- ✅ **Checking rate limiting and DoS protection**
- ✅ **Validating SSL/TLS configuration**
- ❌ **NO malicious testing or exploit attempts**
- ❌ **NO vulnerability exploitation**
- ❌ **NO offensive security tools**

## 🚀 How to Run Security Demos

```bash
# Build the HTTP DSL runner
go build -o bin/http-runner ./runner/http_runner.go

# Run individual security demos
./bin/http-runner scripts/security/security_01_headers.http
./bin/http-runner scripts/security/security_02_auth_validation.http

# Run the complete security suite
./bin/http-runner scripts/security/security_complete.http

# Run with verbose output for detailed security analysis
./bin/http-runner -v scripts/security/security_complete.http

# Validate security scripts without execution
./bin/http-runner --validate scripts/security/security_05_ssl_tls.http
```

## 📚 Security Demo Progression

### 1. Security Headers (`security_01_headers.http`)
**🛡️ Focus:** Critical security headers validation
- ✅ **HSTS (Strict-Transport-Security)** validation
- ✅ **CSP (Content-Security-Policy)** testing
- ✅ **X-Frame-Options** verification
- ✅ **X-Content-Type-Options** validation
- ✅ **CORS headers** security testing
- ✅ **Authentication header** security patterns

**Learn:** How to validate that APIs return proper security headers for client protection

### 2. Authentication Validation (`security_02_auth_validation.http`)  
**🔐 Focus:** Authentication mechanism testing
- ✅ **Bearer token** validation patterns
- ✅ **API key** authentication testing
- ✅ **Multi-factor authentication** simulation
- ✅ **Role-based access control (RBAC)** validation
- ✅ **Session management** security
- ✅ **Token expiry and refresh** patterns

**Learn:** How to test various authentication mechanisms and validate security

### 3. Input Validation (`security_03_input_validation.http`)
**🛡️ Focus:** Input validation and sanitization testing
- ✅ **Length validation** testing
- ✅ **Data type validation** verification  
- ✅ **Special character** handling
- ✅ **JSON structure** validation
- ✅ **Parameter boundary** testing
- ✅ **Input sanitization** patterns

**Learn:** How to validate that APIs properly sanitize and validate all inputs

### 4. Rate Limiting (`security_04_rate_limiting.http`)
**⚡ Focus:** Rate limiting and DoS protection testing
- ✅ **Rate limit compliance** testing
- ✅ **Burst pattern** analysis
- ✅ **Rate limit header** detection
- ✅ **Throttling behavior** validation
- ✅ **Client identification** patterns
- ✅ **Recovery and backoff** testing

**Learn:** How to test rate limiting effectiveness and DoS protection mechanisms

### 5. SSL/TLS Security (`security_05_ssl_tls.http`)
**🔐 Focus:** Transport layer security validation
- ✅ **HTTPS enforcement** testing
- ✅ **Certificate validation** verification
- ✅ **TLS protocol version** checking
- ✅ **Cipher suite** validation
- ✅ **Mixed content prevention** testing
- ✅ **Certificate pinning** patterns

**Learn:** How to validate SSL/TLS configuration and transport security

### 🎯 Complete Security Suite (`security_complete.http`)
**🏆 Focus:** Comprehensive enterprise security validation
- ✅ **ALL security features** in realistic enterprise scenario
- ✅ **Multi-phase security testing** workflow
- ✅ **Compliance validation** (SOC2, ISO27001, GDPR patterns)
- ✅ **Audit trail** generation
- ✅ **Security monitoring** patterns
- ✅ **Incident response** simulation

**Learn:** Real-world application of all security testing features in enterprise environment

## 🛡️ Security Testing Matrix

| Security Control | Demo 1 | Demo 2 | Demo 3 | Demo 4 | Demo 5 | Complete |
|------------------|--------|--------|--------|--------|--------|----------|
| Security Headers | ✅ | | | | ✅ | ✅ |
| Authentication | ✅ | ✅ | | | | ✅ |
| Authorization | | ✅ | | | | ✅ |
| Input Validation | | | ✅ | | | ✅ |
| Rate Limiting | | | | ✅ | | ✅ |
| SSL/TLS Security | | | | | ✅ | ✅ |
| Session Security | | ✅ | | | | ✅ |
| CORS Validation | ✅ | | | | | ✅ |
| Audit Logging | | | | | | ✅ |
| Compliance Check | | | | | | ✅ |

## 🔒 Security Best Practices Demonstrated

### Authentication & Authorization
- **Multi-factor authentication** patterns
- **Role-based access control (RBAC)** implementation
- **Token validation and refresh** mechanisms
- **Session security** best practices

### Input Security
- **Comprehensive input validation** (length, type, format)
- **Special character sanitization** patterns
- **JSON schema validation** techniques
- **Parameter boundary checking**

### Transport Security  
- **HTTPS enforcement** with HSTS
- **Certificate validation** and pinning
- **Strong cipher suites** preference
- **Mixed content prevention**

### API Security
- **Rate limiting** implementation
- **Security headers** configuration
- **CORS policy** validation
- **Error handling** security

### Monitoring & Compliance
- **Security event logging** patterns
- **Audit trail** generation
- **Compliance framework** validation
- **Incident response** procedures

## 🎯 What Each Demo Teaches

| Demo | Security Domain | Key Controls | Difficulty |
|------|----------------|--------------|------------|
| 01 | Headers | HSTS, CSP, CORS | ⭐⭐ |
| 02 | Authentication | Bearer, RBAC, MFA | ⭐⭐⭐ |
| 03 | Input Validation | Sanitization, Types | ⭐⭐⭐ |
| 04 | Rate Limiting | DoS Protection | ⭐⭐⭐ |
| 05 | SSL/TLS | Transport Security | ⭐⭐⭐⭐ |
| Complete | Enterprise | All Controls | ⭐⭐⭐⭐⭐ |

## 🌟 HTTP DSL v3 Security Features

These demos showcase security testing capabilities **NEW and ENHANCED** in v3:

- **🆕 Multi-header Security Testing**: Test complex security header combinations
- **🆕 Block-based Security Flows**: Use multiline if/then/endif for complex security logic  
- **✨ Enhanced JSON Security**: Test APIs with complex JSON containing security-relevant data
- **✨ Advanced Variable Security**: Secure token and session ID handling
- **✨ Loop-based Security Testing**: Repeated security validation tests
- **✨ Comprehensive Error Handling**: Graceful security test failure handling

## 🔍 Security Validation Indicators

When you run these security demos, you should see:

- ✅ **All security tests pass** without errors
- ✅ **Security headers validated** correctly
- ✅ **Authentication mechanisms work** as expected
- ✅ **Input validation functions** properly
- ✅ **Rate limiting** is respected
- ✅ **SSL/TLS connections** are secure
- ✅ **Audit trails** are generated
- ✅ **Compliance checks** pass

If any security demo fails, it may indicate:
- Configuration issues in the API
- Missing security controls
- Need for security improvements
- Environmental connectivity problems

## 🛠️ Troubleshooting Security Tests

**Security tests failing?**
```bash
# Check runner is built correctly
go build -o bin/http-runner ./runner/http_runner.go

# Test with verbose output for details
./bin/http-runner -v scripts/security/security_01_headers.http
```

**Network connectivity issues?**
- Security demos use public APIs (jsonplaceholder.typicode.com, httpbin.org)
- These APIs are designed for testing and should be available 24/7
- Check internet connectivity if tests fail

**Authentication errors?**
- Demos use simulated tokens for testing patterns
- Real APIs may return different responses
- Focus on the testing patterns demonstrated

## ⚠️ Security Testing Ethics

**These demos are for DEFENSIVE security testing only:**

✅ **Ethical Use:**
- Testing your own APIs and systems
- Validating security controls work correctly
- Learning security testing techniques
- Demonstrating security best practices

❌ **Unethical Use:**
- Testing systems you don't own
- Attempting to bypass security controls
- Looking for vulnerabilities to exploit
- Any form of malicious testing

## 📋 Security Compliance

These demos help validate compliance with:
- **SOC 2** - Security controls validation
- **ISO 27001** - Information security management
- **GDPR** - Data protection patterns
- **OWASP** - Web application security best practices
- **NIST** - Cybersecurity framework controls

## 🎊 Security Testing Achievement

**When all security demos pass successfully, you've validated:**

🏆 **Enterprise-Grade Security Controls**
- Complete authentication and authorization
- Comprehensive input validation
- Strong transport layer security  
- Effective rate limiting and DoS protection
- Proper security headers configuration
- Audit trail and compliance readiness

---

🛡️ **Ready for Enterprise Security Testing!** 🛡️

*These security demos prove that HTTP DSL v3 is capable of comprehensive enterprise-grade defensive security testing.*