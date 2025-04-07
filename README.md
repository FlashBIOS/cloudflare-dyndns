# Cloudflare Dynamic DNS

Cloudflare Dynamic DNS (cloudflare-dyndns) is a command‑line tool written in Go that allows you to update and manage Cloudflare DNS records automatically based on your current public IP. It supports IPv4 (“A” records) and IPv6 (“AAAA” records) and provides commands to list and update your DNS records for your Cloudflare zones.

## Features

- **Automatic IP Detection:** Uses the ipify API to determine your current public IP address.
- **DNS Record Updates:** Updates your Cloudflare zone records with the latest public IP.
- **Record Listing:** Lists active DNS records (including A and AAAA records) in a clean, tabulated format.
- **Retries & Backoff:** Implements retry logic with exponential backoff for reliable network operations.

## Prerequisites

- [Go 1.24](https://golang.org/dl/) or later
- A valid Cloudflare account with API access

## Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/yourusername/cloudflare-dyndns.git
   cd cloudflare-dyndns
   ```

2. **Build the project:**

   ```bash
   go build -o cloudflare-dyndns .
   ```

3. **Install:**

   Optionally, move the built binary to a directory in your `$PATH`:

   ```bash
   mv cloudflare-dyndns /usr/local/bin/
   ```

## Configuration

Create a configuration file (the default location for this file is, `~/.cloudflare-dyndns`) in your with the required settings. 

You will need to include details such as your Cloudflare API token. A complete example configuration found in `.cloudflare-dyndns.example`.  

Copy this file to your preferred location, and edit the values.

Make sure to update the placeholder values with your actual configuration details.

## Usage

The tool provides several commands via its CLI. Some common commands include:

- **Show Current IP Address**
  Display your currently assigned public IP address.
  
  ```bash
  cloudflare-dyndns ip
  ```

- **List DNS Records:**  
  Display a list of current DNS records in your Cloudflare zone.

  ```bash
  cloudflare-dyndns list
  ```

- **Update a Record:**  
  Update a specific DNS record with your current public IP. This command is likely the reason you are wanting to use this program. This example uses a custom configuration file. 

  ```bash
  cloudflare-dyndns update --config /etc/cloudflare-dyndns/example.com.config
  ```
  
  Use this command in your crontab or other scheduler to automatically check for IP address changes at an interval.

If you need help with a command, you can typically display the command’s help information:

```bash
cloudflare-dyndns --help
```

## Testing

The project includes tests for core functionality such as fetching your public IP via ipify and handling error cases. To run the tests:

```bash
make test
```

Note that the design isolates any process termination (such as calls to `os.Exit()`) away from testable functions. This means your tests run smoothly without inadvertently stopping any test runs.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request with details about your proposed changes. Be sure to follow Go best practices and include tests for any new functionality.

## License

This project is licensed under a permissive license. See the [LICENSE](LICENSE) file for details.