# HTTP DSL v3 - Ethical Penetration Testing Suite 🔍

**⚠️ CRITICAL: READ [ETHICAL_USE_AGREEMENT.md](./ETHICAL_USE_AGREEMENT.md) BEFORE USING**

A comprehensive collection of **ethical penetration testing** scripts for HTTP APIs using HTTP DSL v3. These demos are designed for **authorized security testing only** - testing systems you own or have explicit permission to test.

## 🚨 **ETHICAL USE ONLY**

### ✅ **AUTHORIZED USE:**
- Your own APIs and applications
- Development/testing environments
- Authorized penetration testing with written permission
- Educational security training environments
- Public testing platforms (DVWA, WebGoat, etc.)

### ❌ **PROHIBITED USE:**
- Third-party systems without permission
- Production systems without authorization  
- Any malicious or illegal activities
- Unauthorized vulnerability scanning

## 📋 **LEGAL REQUIREMENTS**

**Before using these demos, ensure you have:**
1. **Written authorization** from system owners
2. **Legal compliance** in your jurisdiction
3. **Documented scope** and limitations
4. **Incident response plan** for findings
5. **Responsible disclosure process**

## 🎯 **Penetration Testing Demos**

### 1. SQL Injection Testing (`pentest_01_sql_injection.http`)
**💉 Focus:** Testing for SQL injection vulnerabilities
- ✅ **Basic injection patterns** (OR '1'='1, UNION SELECT)
- ✅ **Time-based blind injection** (SLEEP, WAITFOR)
- ✅ **Boolean-based blind injection** (true/false conditions)
- ✅ **UNION-based injection** (data extraction)
- ✅ **Second-order injection** (stored payload execution)
- ✅ **Error-based injection** (database-specific errors)

**Learn:** How to systematically test for SQL injection vulnerabilities using various techniques

### 2. Cross-Site Scripting Testing (`pentest_02_xss_testing.http`)  
**🌐 Focus:** Testing for XSS vulnerabilities
- ✅ **Reflected XSS** (basic, img, svg payloads)
- ✅ **Stored/Persistent XSS** (comments, posts)
- ✅ **DOM-based XSS** (hash, javascript: URLs)
- ✅ **Filter evasion** (case, encoding, character codes)
- ✅ **Context-specific XSS** (HTML, JS, JSON contexts)
- ✅ **Advanced techniques** (polyglot, mutation, template injection)

**Learn:** How to test for XSS vulnerabilities across different contexts and bypass common filters

## 🛠️ **How to Run Penetration Tests**

```bash
# Build the HTTP DSL runner
go build -o bin/http-runner ./runner/http_runner.go

# Read the ethical use agreement first
cat scripts/penetration_testing/ETHICAL_USE_AGREEMENT.md

# Run SQL injection testing (authorized systems only)
./bin/http-runner scripts/penetration_testing/pentest_01_sql_injection.http

# Run XSS testing (authorized systems only)
./bin/http-runner scripts/penetration_testing/pentest_02_xss_testing.http

# Run with verbose output for detailed analysis
./bin/http-runner -v scripts/penetration_testing/pentest_01_sql_injection.http
```

## 📊 **Penetration Testing Matrix**

| Vulnerability Type | Test Coverage | Techniques | Real-World Relevance |
|-------------------|---------------|------------|---------------------|
| SQL Injection | ✅ Comprehensive | 6+ attack vectors | ⭐⭐⭐⭐⭐ Critical |
| XSS | ✅ Comprehensive | 6+ attack contexts | ⭐⭐⭐⭐⭐ Critical |
| CSRF | 🔄 Planned | Token validation | ⭐⭐⭐⭐ High |
| SSRF | 🔄 Planned | Internal requests | ⭐⭐⭐⭐ High |
| XXE | 🔄 Planned | XML parsing | ⭐⭐⭐ Medium |
| Path Traversal | 🔄 Planned | Directory access | ⭐⭐⭐ Medium |

## 🔍 **What These Tests Detect**

### SQL Injection Testing Detects:
- **Classic injection** vulnerabilities
- **Blind injection** (time-based, boolean-based)
- **Union-based** data extraction possibilities
- **Error-based** information disclosure
- **Second-order** injection risks
- **Database-specific** vulnerabilities

### XSS Testing Detects:
- **Reflected XSS** in parameters and headers
- **Stored XSS** in user-generated content
- **DOM-based XSS** in client-side code
- **Filter bypass** opportunities
- **Context-specific** injection points
- **Advanced XSS** vectors

## 🛡️ **Security Best Practices Demonstrated**

### For SQL Injection Prevention:
- **Parameterized queries** / prepared statements
- **Input validation** and sanitization
- **Least privilege** database permissions
- **Web Application Firewall** (WAF) rules
- **Database activity monitoring**
- **Regular security code reviews**

### For XSS Prevention:
- **Input validation** and output encoding
- **Content Security Policy** (CSP) headers
- **Context-aware encoding** (HTML, JS, CSS, URL)
- **HttpOnly cookie flags**
- **X-XSS-Protection headers**
- **Secure development practices**

## 📚 **Educational Value**

These penetration testing demos teach:

1. **Vulnerability Assessment Methodology**
   - Systematic testing approaches
   - Different attack vectors for each vulnerability type
   - How to identify and classify findings

2. **Security Testing Techniques**
   - Manual testing vs automated scanning
   - Context-specific testing approaches
   - Filter evasion and bypass techniques

3. **Risk Assessment Skills**
   - Understanding impact and likelihood
   - Prioritizing security findings
   - Communicating risks to stakeholders

4. **Defensive Security Mindset**
   - How attackers think and operate
   - Common developer security mistakes
   - Effective countermeasures and controls

## 🔧 **Recommended Testing Environments**

### **✅ Safe Testing Platforms:**
- **DVWA** (Damn Vulnerable Web Application)
- **WebGoat** (OWASP)
- **bWAPP** (Buggy Web Application)
- **Mutillidae** (OWASP)
- **SQLi-Labs** (SQL injection practice)
- **XSS Hunter** (XSS detection platform)

### **✅ Professional Tools Integration:**
These HTTP DSL demos complement professional tools:
- **Burp Suite** (Commercial web app scanner)
- **OWASP ZAP** (Free security scanner)
- **SQLMap** (SQL injection tool)
- **Nuclei** (Fast vulnerability scanner)
- **Nessus** (Enterprise vulnerability scanner)

## 🎯 **Real-World Application**

These demos prepare you for:

### **Authorized Penetration Testing**
- Web application security assessments
- API security testing
- Red team exercises
- Compliance testing (PCI DSS, OWASP Top 10)

### **Security Development**
- Secure code reviews
- Security testing integration in CI/CD
- Developer security training
- Vulnerability remediation validation

### **Bug Bounty Programs**
- Responsible disclosure practices
- Systematic vulnerability hunting
- Report writing and communication
- Ethical hacking methodologies

## ⚠️ **Important Limitations**

**What these demos ARE:**
✅ Educational security testing patterns
✅ Vulnerability assessment techniques
✅ Security awareness training tools
✅ Authorized penetration testing aids

**What these demos are NOT:**
❌ Complete vulnerability scanners
❌ Replacement for professional security tools
❌ Automated exploitation frameworks
❌ Tools for unauthorized testing

## 📋 **Responsible Disclosure Process**

If you find real vulnerabilities using these techniques:

1. **Stop testing** immediately upon discovery
2. **Document** the vulnerability thoroughly
3. **Report** to the system owner/security team
4. **Wait** for acknowledgment and remediation timeline
5. **Don't disclose** publicly until patched
6. **Follow up** appropriately

## 🏆 **Certification and Training Value**

These demos support preparation for:
- **CEH** (Certified Ethical Hacker)
- **OSCP** (Offensive Security Certified Professional)
- **GWEB** (GIAC Web Application Penetration Tester)
- **CSSLP** (Certified Secure Software Lifecycle Professional)
- **Security+** (CompTIA Security Certification)

## 🔒 **Final Security Reminders**

### **Always Remember:**
1. **Permission first** - Never test without authorization
2. **Document everything** - Keep detailed records
3. **Follow the law** - Know your legal boundaries  
4. **Minimize impact** - Use least intrusive methods
5. **Report responsibly** - Help fix, don't exploit

### **Ethical Hacker's Oath:**
*"I will use my security knowledge to protect and defend, never to harm or exploit. I will respect privacy, follow the law, and contribute to a safer digital world."*

---

🛡️ **Use These Powers for Good!** 🛡️

*Ethical penetration testing makes the internet safer for everyone. Test responsibly, learn continuously, and help build a more secure world.*