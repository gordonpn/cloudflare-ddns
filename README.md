# Cloudflare DDNS

## Usage

### Configuration

| Name        | Description                                                                                                                                   | Type         | Required |
| ----------- | --------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | -------- |
| API_TOKEN   | Your Cloudflare API token: <https://dash.cloudflare.com/profile/api-tokens>                                                                   | String       | x        |
| ZONE_ID     | Your Cloudflare Zone ID                                                                                                                       | String       | x        |
| RECORD_NAME | The name of the A record you want to update. (e.g. gordon-pn.com)                                                                             | String       | x        |
| RECORD_TTL  | Defaults to 1 (auto). [More info.](https://support.cloudflare.com/hc/en-us/articles/360017421192-Cloudflare-DNS-FAQ#h_2kCxAtTEHDevfWwIWxf1m0) | Number       |          |
| APP_ENV     | The environment the program is running in                                                                                                     | String       |          |
| HC_URL      | [Healthchecks.io URL](https://healthchecks.io/)                                                                                               | String (URL) | x        |

## Related Projects

- [asuswrt-merlin-ddns-cloudflare](https://github.com/alphabt/asuswrt-merlin-ddns-cloudflare)
- [cloudflare-ddns](https://github.com/timothymiller/cloudflare-ddns)
