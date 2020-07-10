Get lines of code for GitHub repository or user or organization

## Usage

Download a binary from GitHub releases: https://github.com/g4s8/getloc/releases

Or build it from sources (Go 1.14 required):
```
git clone --depth=1 https://github.com/g4s8/getloc.git
cd getloc
go build
```

Run `getloc` with repository or organization as argument:
```
getloc artipie/artipie # get LOC for repository
# or
getloc artipie # get LOC for all repos in organization
```
