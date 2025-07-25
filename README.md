# mcp fetch and command tool

## Installation

1. Get the latest [release](https://github.com/tiechui1994/mcp) and put it in your `$PATH` or somewhere you can easily access.

2. Or if you have Go installed, you can build it from source:

```sh
go install -v https://github.com/tiechui1994/mcp@latest
```

## Tools

### Schema Tools

1. `fetch`

    - ${mcp.tool.fetch.desc}
    - Parameters: 
        - `url`: Required, fetch url 
    - Returns: Fetch url GET request result

2. `cmd`

    - ${mcp.tool.cmd.desc}
    - Parameters:
        - `cmd`: Required, local exec command
    - Returns: Exec command result


## License

Apache