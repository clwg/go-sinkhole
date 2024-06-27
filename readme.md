# go-sinkhole

This is a Golang-based sinkhole server designed to intercept and log unwanted or malicious network traffic for analysis and mitigation purposes.  It sinkhole both TCP and UDP traffic and listen on multiple ports.

## Usage

```bash
go run cmd.go -filenamePrefix <prefix> -logDir <directory> -maxLines <lines> -rotationTime <minutes> -protocol <protocol> -ports <ports>
```
