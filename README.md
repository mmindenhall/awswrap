# awswrap

A lightweight Go replacement for [aws2-wrap](https://github.com/linaro-its/aws2-wrap), providing a subset of its functionality as a standalone binary.

## Why awswrap?

aws2-wrap is a Python tool that wraps commands with AWS SSO credentials. It works well, but being a Python package means it requires a Python environment and can be awkward to use with virtualenvs — if you install aws2-wrap globally but activate a project virtualenv, the `aws2-wrap` command may no longer be on your PATH.

awswrap solves this by providing the same core functionality as a single Go binary with no runtime dependencies.

## Installation

```sh
go install github.com/mmindenhall/awswrap@latest
```

Or build from source:

```sh
git clone https://github.com/mmindenhall/awswrap.git
cd awswrap
go build -o awswrap .
```

## Usage

awswrap reads your `~/.aws/config` profiles (including SSO and assume-role chains) and resolves temporary credentials, then either exports them or passes them to a wrapped command.

### Wrap a command

```sh
awswrap --profile my-profile aws s3 ls
```

The command receives AWS credentials via environment variables. This is the default mode when no other flag is specified.

### Wrap a command with --exec

```sh
awswrap --profile my-profile --exec "aws s3 ls | grep mybucket"
```

Use `--exec` when your command includes shell features like pipes or redirects. The command string is passed to `sh -c` (or `cmd /C` on Windows).

### Export credentials

Running with `--export` prints shell statements that set AWS credential environment variables:

```sh
awswrap --profile my-profile --export
# Output:
# export AWS_ACCESS_KEY_ID=AKIA...
# export AWS_SECRET_ACCESS_KEY=...
# export AWS_SESSION_TOKEN=...
```

To apply them to your current shell session, wrap the call with `eval`:

```sh
eval $(awswrap --profile my-profile --export)
```

On PowerShell, the output uses `$ENV:` syntax instead.

### Profile resolution

- `--profile` defaults to `$AWS_PROFILE`, then `$AWS_DEFAULT_PROFILE`, then `default`
- Supports `source_profile` chains (for role assumption)
- Supports `sso_session` references
- Reads `$AWS_CONFIG_FILE` or defaults to `~/.aws/config`

### Flags

| Flag | Description |
|------|-------------|
| `--profile <name>` | AWS config profile to use |
| `--export` | Print credential export statements |
| `--exec <command>` | Run command string through the system shell |
| `-v`, `--version` | Print version |

## Differences from aws2-wrap

awswrap intentionally supports only a subset of aws2-wrap's features:

- **Supported:** command wrapping, `--export`, `--exec`, `--profile`, SSO login, assume-role chains
- **Not supported:** `--generate`, `--generatestdout`, `--process`, `--outprofile`, `--configfile`, `--credentialsfile`

The dropped features involve writing credentials to AWS config/credentials files or outputting credential_process JSON, which are niche use cases that the supported modes cover adequately.
