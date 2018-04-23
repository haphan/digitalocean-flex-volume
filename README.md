
# DigitalOcean Flex Volume

This is a kubernetes FlexVolume driver for DigitalOcean.

## Configuration

Copy the plugin binary to the kubernetes volume plugin directory at every node, including master nodes.
That binary will be used by kubelet and kube-configuration-manager.

The plugin needs to use the DigitalOcean token which should be configured in a file or an environment variable:

| Environment Variable              | default                    | Description                                                                                                                                   |
|-----------------------------------|----------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| `DIGITALOCEAN_TOKEN_FILE_PATH` | /etc/kubernetes/digitalocean.json | Complete path to the file containing the DigitalOcean Token     |
| `DIGITALOCEAN_TOKEN`       |                 | The token file takes precedence over this environment variable |

## Troubleshoot

If your `kubelet` or `kubernetes-controller-manager` is running as a container, make sure that:
 - the plugin path is being mapped in the container
 - if using environment variables to configure the token or token file path, make sure those variables are created in the container
 - if the `DIGITALOCEAN_TOKEN_FILE_PATH` is being used, check that the path to the token file exists in the container
