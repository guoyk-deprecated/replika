# replika

replicate container images by list

## Usage

* Environment Variable
  * `DOCKERCONFIG_BASE64`, optional, base64 encoded docker config.json
* Arguments
  * `-f`, images list file
  * `-c`, concurrency, default to 5
  * `-src`, source registry, for example, `ghcr.io/guoyk93/acicn`
  * `-dst`, destination registries, comma seperated, for example `acicn,ccr.ccs.tencentyun.com/acicn`
  * `-pull`, pull the image from source registry
  * `-push`, push the image to destination registries
  * `-docker-config`, override the default docker config directory

## Donation

View <https://guoyk.net/donation>

## Credits

Guo Y.K., MIT License
