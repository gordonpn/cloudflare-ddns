# Cloudflare DDNS

## Description

Keep your Cloudflare type A DNS record up to date with your dynamic IP address.

---

[![Project Status: Active](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#active)
[![Build Status](https://drone.gordon-pn.com/api/badges/gordonpn/cloudflare-ddns/status.svg)](https://drone.gordon-pn.com/gordonpn/cloudflare-ddns)
![License](https://badgen.net/github/license/gordonpn/hot-flag-deals)

[![Buy Me A Coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/gordonpn)

## Motivation

There exists a pain point between self-hosting projects and being a subscriber to an internet service provider with a dynamic IP address.

This is a Dynamic DNS (DDNS) solution to keep an ever-changing IP address updated with a Cloudflare A record.

Written in Go, because why not.

## Usage

There are a couple ways to run the program.

### Method 1

Run as a one-off Go script: `go run ./cmd/cloudflare-ddns/main.go`.

### Method 2

Run the Docker container: `docker run ghcr.io/gordonpn/cloudflare-ddns:stable`

Note that running with the container will require environment variables to be set in the shell.

#### Flags

There is currently one flag available. `-periodic` which blocks the program from terminating and updates the IP address continuously every 5 minutes. This is useful if you want to "run and forget".

#### Cron

An alternative way is to create a recurring cron task with your own specified time.

E.g. every hour: <https://crontab.guru/every-1-hour>.

### Configuration

Either copy the `.env.example` to `.env` or inject the environment variables however you want and fill in the variables below.

| Name        | Description                                                                                                                                   | Type         | Required |
| ----------- | --------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | -------- |
| API_TOKEN   | Your Cloudflare API token: <https://dash.cloudflare.com/profile/api-tokens>                                                                   | String       | x        |
| ZONE_ID     | Your Cloudflare Zone ID                                                                                                                       | String       | x        |
| RECORD_NAME | The name of the A record you want to update. (e.g. gordon-pn.com)                                                                             | String       | x        |
| RECORD_TTL  | Defaults to 1 (auto). [More info.](https://support.cloudflare.com/hc/en-us/articles/360017421192-Cloudflare-DNS-FAQ#h_2kCxAtTEHDevfWwIWxf1m0) | Number       |          |
| APP_ENV     | The environment the program is running in                                                                                                     | String       |          |
| HC_URL      | [Healthchecks.io URL](https://healthchecks.io/)                                                                                               | String (URL) | x        |

## Getting Started with Development

### Prerequisites

- Golang v1.16+
- Docker v20.10+

## Related Projects

- [asuswrt-merlin-ddns-cloudflare](https://github.com/alphabt/asuswrt-merlin-ddns-cloudflare)
- [cloudflare-ddns](https://github.com/timothymiller/cloudflare-ddns)

## Support

You may open an issue for discussion.

## Authors

[@gordonpn](https://github.com/gordonpn)

## License

[MIT License](./LICENSE)
