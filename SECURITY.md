# Security Policy

## Supported Versions

We actively support the following versions of the Teamwork.com API Go SDK with
security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | ✅                 |
| < 1.0   | ❌                 |

## Reporting a Vulnerability

The Teamwork team takes security seriously. If you discover a security
vulnerability in this SDK, please report it responsibly.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report security vulnerabilities to:

- **Email**: security@teamwork.com
- **Subject**: [Security] Teamwork API Go SDK - Brief description of vulnerability

### What to Include

When reporting a vulnerability, please include:

1. **Description**: A clear description of the vulnerability
2. **Impact**: The potential impact of the vulnerability
3. **Reproduction**: Step-by-step instructions to reproduce the issue
4. **Affected Versions**: Which versions of the SDK are affected
5. **Environment**: Go version, OS, and any other relevant environment details
6. **Proof of Concept**: If applicable, include a minimal proof of concept

### Response Timeline

We will acknowledge receipt of your vulnerability report within **48 hours** and
will strive to:

- Provide an initial assessment within **5 business days**
- Keep you informed of our progress toward a fix
- Notify you when the vulnerability is resolved

### Responsible Disclosure

We kindly ask that you:

- Allow us reasonable time to address the vulnerability before public disclosure
- Do not access, modify, or delete data belonging to others
- Do not perform actions that could negatively impact Teamwork users or services
- Do not publicly disclose the vulnerability until we have had a chance to address it

## Security Best Practices

When using this SDK, we recommend following these security best practices:

### API Key Management

- **Never commit API keys to version control**
- Store API keys securely using environment variables or secure credential stores
- Rotate API keys regularly
- Use the minimum required permissions for your integration

```go
// ✅ Good: Use environment variables
apiKey := os.Getenv("TEAMWORK_API_KEY")

// ❌ Bad: Hardcoded API keys
apiKey := "tk_live_12345..."
```

### Authentication Methods

- **OAuth2**: Recommended for applications where users authenticate themselves
- **Bearer Token**: Suitable for server-to-server integrations
- **Basic Auth**: Legacy method, use only when necessary

### Network Security

- Always use HTTPS endpoints (this SDK enforces HTTPS by default)
- Implement proper timeout and retry mechanisms
- Consider rate limiting in your application

### Error Handling

- Avoid logging sensitive information in error messages
- Implement proper error handling to prevent information disclosure

```go
// ✅ Good: Generic error handling
if err != nil {
  log.Printf("API request failed: %v", err.Error())
  return
}

// ❌ Bad: Logging potentially sensitive data
if err != nil {
  log.Printf("Failed with token %s: %v", apiKey, err)
  return
}
```

### Data Handling

- Only request the data you need
- Implement proper data validation
- Follow data retention policies
- Ensure secure data transmission and storage

## Dependencies

This SDK has minimal dependencies to reduce the attack surface:

- **golang.org/x/sys**: System-level operations (OS-specific functionality)

We regularly monitor our dependencies for known vulnerabilities and update them
as needed.

## Security Features

### Built-in Security Measures

- **HTTPS Enforcement**: All API calls are made over HTTPS
- **Context Support**: Built-in support for request timeouts and cancellation
- **Input Validation**: Request parameters are validated before sending
- **Error Sanitization**: Error messages are sanitized to prevent information leakage

### Recommended Usage Patterns

```go
// Use context with timeout for all requests
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Make API calls with context
response, err := client.GetProjectsWithContext(ctx, params)
if err != nil {
  // Handle error appropriately
  return err
}
```

## Vulnerability Disclosure Policy

Once a security vulnerability is resolved:

1. We will publish a security advisory on GitHub
2. Release notes will include security fix information
3. We may coordinate with you on public disclosure timing
4. Credit will be given to the reporter (unless they prefer to remain anonymous)

## Security Resources

- [Teamwork Security](https://www.teamwork.com/security)
- [Teamwork API Documentation](https://apidocs.teamwork.com/)
- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)

## Contact

For non-security related questions about this SDK:
- Open an issue on [GitHub](https://github.com/teamwork/twapi-go-sdk/issues)
- Check our [documentation](https://apidocs.teamwork.com/)

For security-related inquiries: security@teamwork.com

---

*This security policy is effective as of July 10, 2025 and may be updated from time to time.*