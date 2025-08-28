# stagen

static git web viewer. Generates static HTML pages for git repositories.

## Usage

Generate static pages for a repository:
```
./stagen -repo /path/to/git/repo -out repos/myrepo -name "My Repo" -desc "Description"
```

After generating repos, your directory structure will look like:
```
├── repos/
│   ├── index.html          # Main repository index
│   ├── style.css           # Shared CSS file
│   ├── myrepo/
│   │   ├── index.html      # Repository files view
│   │   ├── log.html        # Commit log
│   │   ├── commits.html    # Commits table
│   │   ├── refs.html       # Branches/tags
│   │   ├── readme.html     # README display
│   │   ├── file/           # Individual files
│   │   └── commit/         # Individual commits
│   └── another/
│       └── ...
├── stagen                  # Bin
└── *.tmpl                  # Template files
```

## Features

- Single file implementation
- No dependencies beyond Go stdlib
- Diff syntax highlighting
- Multiple repository index
- Uses Go's template approach

## Build

```
go build stagen.go
```
