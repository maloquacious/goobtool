# Quickstart Guide

## Getting Started

```bash
# Create and initialize the store
app db create

# Start the server (defaults to ports 8080 and 8383)
app serve --port 8080 --admin-port 8383

# Check server status (via local admin API)
app server status

# Perform a graceful restart
app server restart

# Shut down the server
app server shutdown
```

## Additional Notes

- Ensure you have built the application using `make build` or the equivalent Go build command.
- The admin API is accessible only on the loopback interface for security.
- Refer to the main README.md for more details on configuration and features.
