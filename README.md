# gws - Google Cloud Workstation Utils

A command-line tool to manage Google Cloud Workstations.

## Description

`gws` is a utility to simplify the management of Google Cloud Workstations. It provides commands to start, stop, and connect to your workstation, as well as manage your configuration.

## Installation

You can install `gws` using `go install`:

```bash
go install github.com/bakito/gws@latest
```

## Usage

### Commands

- `gws setup`: Create a new or update the config.yaml and create a context configuration using an interactive terminal setup wizard.
- `gws start [context]`: Start the workstation for the given or current context.
- `gws stop [context]`: Stop the workstation for the given or current context.
- `gws restart [context]`: Restart the workstation for the given or current context.
- `gws up`: Uploads files and directories to the workstation as defined in the context configuration.
- `gws tunnel [context]`: Create an SSH tunnel to the workstation.
- `gws patch`: Patch local files as defined in the `filePatches` configuration.
- `gws ctx [context]`: Switch the current context. If no context is provided, an interactive selection is shown.
  - `--current`: Print the current active context.

### Global Flags

- `--config, -c`: Path to the configuration file (default: `~/.gws/config.yaml`).
- `--ctx`: The context to use.

## Configuration

`gws` is configured using a YAML file (default: `~/.gws/config.yaml`). You can use the `gws setup` command to create an initial configuration.

The configuration file can contain multiple contexts. Each context defines the connection details for a specific workstation.

### `config.yaml` example

```yaml
current-context: my-workstation
contexts:
  my-workstation:
    host: localhost
    port: 2222
    user: user
    private-key-file: /path/to/your/private/key
    known-hosts-file: /path/to/your/known_hosts
    gcloud:
      project: my-project
      region: a-region
      cluster: my-cluster
      config: my-workstation-config
      name: my-workstation
    dirs:
    - path: /home/user/.ssh
      permissions: "0700"
    files:
    - source-path: /path/to/your/file
      path: /home/user/file
      permissions: "0644"
```

### Configuration Options

- `current-context`: The name of the currently active context.
- `contexts`: A map of contexts.
  - `<context-name>`:
    - `host`: The hostname or IP address of the workstation.
    - `port`: The port to connect to.
    - `user`: The username to use for the SSH connection.
    - `private-key-file`: The path to the private key for the SSH connection.
    - `known-hosts-file`: The path to the known hosts file for the SSH connection.
    - `gcloud`: The Google Cloud configuration.
      - `project`: The Google Cloud project.
      - `region`: The Google Cloud region.
      - `cluster`: The Google Cloud cluster.
      - `config`: The workstation configuration.
      - `name`: The name of the workstation.
    - `dirs`: A list of directories to create on the workstation.
      - `path`: The path of the directory.
      - `permissions`: The permissions of the directory.
    - `files`: A list of files to upload to the workstation.
      - `source-path`: The path of the local file.
      - `path`: The path of the remote file.
      - `permissions`: The permissions of the remote file.
- `file-patches`: A map of file patches.
  - `<patch-name>`:
    - `file`: The path of the file to patch.
    - `patches`: A list of patches to apply.
      - `old`: The old string to replace.
      - `new`: The new string to replace the old string with.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License.
