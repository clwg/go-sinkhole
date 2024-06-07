# go-sinkhole

This is a Golang-based sinkhole server designed to intercept and log unwanted or malicious network traffic for analysis and mitigation purposes.  It sinkhole both TCP and UDP traffic and listen on multiple ports.

## Usage

```bash
go run cmd.go -filenamePrefix <prefix> -logDir <directory> -maxLines <lines> -rotationTime <minutes> -protocol <protocol> -ports <ports>
```

- <prefix>: The prefix for log filenames. Default is "sinkholeserver".
- <directory>: The directory where log files will be stored. Default is "./logs".
- <lines>: The maximum number of lines per log file. Default is 100000.
- <minutes>: The log rotation time in minutes. Default is 60.
- <protocol>: The protocol to use. Must be either "tcp" or "udp".
- <ports>: A comma-separated list of ports or port ranges (e.g., "8000,8001-8005").