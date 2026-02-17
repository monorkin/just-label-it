# Just Label It

Just Label It, aka `jli`, is a CLI tool for quickly labeling images, video and audio files.

It's intentionally simple, just a single binary with an embedded web UI.

- **Supported formats** — JPEG, PNG, GIF, WebP, AVIF, SVG, TIFF, BMP, MP4, WebM, MKV, AVI, MOV, MP3, WAV, OGG, FLAC, AAC, and more
- **One directory = one project** — run `jli` in any directory to label all the media files within it
- **Local storage** — everything saved to `jli.db` in the target directory
- **Keyboard navigation** — Left/Right arrow keys to move between files
- **Labels** — tag files with reusable labels, autocomplete from existing labels
- **Descriptions** — free-text description per file, auto-saved as you type
- **Keyframes** — for video and audio files, mark points in time with their own labels and descriptions

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/monorkin/just-label-it/main/install.sh | sh
```

Or with Go:

```bash
go install github.com/monorkin/just-label-it@latest
```

## Usage

```bash
# Label files in the current directory (opens browser automatically)
jli

# Label files in a specific directory
jli ~/photos

# Start the server without opening a browser
jli serve --port 8080 ~/photos
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--bind` | `127.0.0.1` | Address to bind the server to |
| `--port` | `0` (auto) | Port to listen on |

## Building

To build from source, ensure you have Go 1.25+ installed, then run:

```bash
make build    # outputs ./bin/jli
```

### Cross-compilation

Build for all supported platforms:

```bash
make build-all    # outputs to ./dist/
```

Targets: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
