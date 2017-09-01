
# Digital Ocean Flex Volume

This is a kubernetes FlexVolume driver for Digital Ocean.

# Build
// TODO

# Configuration

Copy the plugin binary to the kubernetes volume plugin directory at every node, including master.
That binary will be used by kubelet and kube-configuration-manager.

| Environment Variable              | default                    | Description                                                                                                                                   |
|-----------------------------------|----------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| `DIGITALOCEAN_TOKEN_FILE_PATH` | will look at ./token | Complete path to the file containing the Digital Ocean Token     |
| `DIGITALOCEAN_TOKEN`       | ""                | If the token file exists, this variable won't be read |


# Bootstrap
