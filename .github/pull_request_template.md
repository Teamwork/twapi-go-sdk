## 👋 Thanks for contributing to the Teamwork Go SDK!

We're excited to see what you've built! This template will help us understand
your changes and get them merged faster.

**Pro tip:** The more details you provide, the quicker we can review and merge your PR! 🎯

---

## 🎯 What type of change is this?
<!-- Pick the one that best describes your PR (following our CONTRIBUTING.md guidelines) -->
- [ ] Feature: New functionality
- [ ] Fix: Bug fixes
- [ ] Docs: Documentation changes
- [ ] Test: Test additions/changes
- [ ] Refactor: Code refactoring
- [ ] Enhancement: Improvements to existing features
- [ ] Chore: Maintenance tasks

## 📋 What does this PR do?
<!-- Give us the elevator pitch! What problem does this solve or what feature does it add? -->

Example: "This PR adds OAuth2 refresh token support to the session package,
allowing long-running applications to automatically refresh expired tokens
without user intervention."

## 🤔 Why is this change needed?
<!-- Help us understand the context and motivation -->

Example: "Users reported that their applications would fail after 1 hour when
tokens expired. This was causing production issues for several customers who run
batch processes."

## 🔨 What changes did you make?
<!-- List the main changes (you can use bullet points) -->

• Added `RefreshToken()` method to `OAuth2Session`
• Implemented automatic token refresh in `makeRequest()`
• Added new error types for token refresh failures
• Updated examples with refresh token usage

## 🧪 How did you test this?
<!-- Describe your testing strategy (see CONTRIBUTING.md for testing guidelines) -->

Testing checklist:
• Unit tests: Added/updated tests for new functionality
• Coverage: Ran `go test -v -cover ./...` 
• Integration: Tested with `TWAPI_SERVER=https://yourdomain.teamwork.com/ TWAPI_TOKEN=your_token go test -v ./...`
• Linting: Verified with `golangci-lint -c .golangci.yml run ./...`
• Manual testing: Describe any manual verification steps
• Examples: Updated/tested relevant examples in examples/

## 🔒 Security Considerations
<!-- Please review our SECURITY.md and confirm these security aspects -->
- [ ] My changes do not introduce security vulnerabilities
- [ ] I have not hardcoded API keys, tokens, or sensitive data
- [ ] Error handling does not expose sensitive information
- [ ] Network communications use HTTPS (enforced by SDK)
- [ ] Input validation is properly implemented for new endpoints

## 💥 Breaking Changes
<!-- Does this PR introduce any breaking changes? -->
- [ ] This PR introduces breaking changes
- [ ] I have updated the documentation to reflect breaking changes
- [ ] I have updated the version number appropriately

### 💥 Breaking Change Details
<!-- If you checked breaking changes above, please describe them -->

Example: "The `Login()` method now returns an additional error parameter for
refresh token failures. Update your error handling code."

## ✅ Contribution Checklist
<!-- Please confirm you've followed our contribution guidelines (CONTRIBUTING.md) -->
- [ ] My code follows Go best practices and project style guidelines
- [ ] I have performed a self-review of my code
- [ ] My code has proper documentation (doc comments for exported functions/types)
- [ ] I have added tests that prove my fix/feature works
- [ ] New and existing unit tests pass locally with `go test -v ./...`
- [ ] I have run `go fmt` and `go vet` on my code
- [ ] I have updated documentation/examples as needed
- [ ] My changes generate no new warnings or errors

## 🤝 Community Values
<!-- Please confirm you've embraced our community values (CODE_OF_CONDUCT.md) -->
- [ ] I have been respectful and constructive in all interactions
- [ ] My contribution aligns with Teamwork's values of excellence and kindness
- [ ] I have followed the Go community's "Gopher values"

## 📎 Additional Context
<!-- Anything else we should know? Links, screenshots, related issues, etc. -->

• Related to issue #123
• Fixes compatibility with Teamwork API version X.X
• Screenshots or examples of new functionality
• Links to relevant Teamwork API documentation
• Performance benchmarks (if applicable)
• Migration notes for users (if breaking changes)
• Special deployment or environment considerations

---

## 🎉 Ready for Review!

Thank you for contributing to the Teamwork Go SDK! Your contribution helps make
project management better for teams worldwide. 🌍

### What happens next:

1. **🤖 Automated checks** - Tests, linting, and security scans
2. **👥 Code review** - Our maintainers will review following our values of kindness and excellence
3. **💬 Collaboration** - We may ask questions or suggest improvements (all in the spirit of learning!)
4. **🚀 Merge and celebrate** - Your code becomes part of the SDK!

### Our Review Values:
- **🤝 We're open and trustworthy** - Transparent feedback and honest communication
- **🔥 We're passionate** - Enthusiasm for great code and user experience  
- **👥 Community first** - Decisions that benefit all SDK users
- **⭐ Excellence** - High standards with patience for the learning process
- **💙 Kindness** - Respectful, constructive, and supportive interactions

Questions? Check our [CONTRIBUTING.md](CONTRIBUTING.md) or ask in the discussion! 

**Thanks for making Teamwork better!** 🙌